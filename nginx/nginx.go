package nginx

import (
    "os"
    "os/exec"
    "sync"
    "bytes"
    "path"
    "bufio"
    "strings"
    "strconv"
    "fmt"
    "text/template"
    "errors"
    "github.com/astutesparrow/lbagent/model"
    "github.com/astutesparrow/lbagent/util"
)

var (
    fileMutex sync.Mutex
    buffer bytes.Buffer
)

const (
    NGX_CONF_PATH = "/etc/nginx"
    NO_SUCH_DOMAIN = "NO_SUCH_DOMAIN"
    INTERNAL_ERROR = "INTERNAL_ERROR"
    SERVER = "server {{.Ip}}:{{.Port}} weight={{if .Weight}}{{.Weight}}{{else}}10{{end}} max_fails=2 fail_timeout={{if .FailTimeout}}{{.FailTimeout}}{{else}}30{{end}}s"
)

func Delete(domain string, dir string) error {
    if domain == "" {
        return errors.New("empty domain")
    }
    
    fileMutex.Lock()
    defer fileMutex.Unlock()

    if dir == "" {
        dir = NGX_CONF_PATH
    }

    file := path.Join(dir, "conf.d", fmt.Sprintf("%s.conf", domain))
    filebak := path.Join(dir, "conf.d", fmt.Sprintf("%s.conf.bak", domain))

    exist := util.Exist(filebak)
    if exist {
        os.Remove(filebak)
    }
    
    err := os.Rename(file, file + ".bak")
    if err != nil {
        util.Log("rename app conf file error, name: " + file)
        return errors.New("delete file failed") 
    }
    
    return nil
}

func Servers(domain string, confpath string) []model.RealServer {
    if domain == "" {
        return nil
    }

    if confpath == "" {
        confpath = NGX_CONF_PATH
    }

    fileMutex.Lock()
    defer fileMutex.Unlock()

    file := path.Join(confpath, "conf.d", fmt.Sprintf("%s.conf", domain))
    
    // read all server lines in upstream block
    fil, err := os.Open(file)
    if err != nil {
        util.Log("open app conf file error, name: " + file)
        return nil
    }
    defer fil.Close()

    lines := make([]string, 0)
    match := fmt.Sprintf("backend_%s", domain)
    flag := false

    scanner := bufio.NewScanner(fil)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(line, "}") {
            flag = false
            break
        }
        if flag {
            lines = append(lines, line)
        }
        if strings.Contains(line, "upstream") && strings.Contains(line, match) {
            flag = true
        }
    }

    res := make([]model.RealServer, 0)
    for _, line := range lines {
        tline := strings.Trim(line, " ")
        if tline == "" {
            break
        }
        var s1, s2, s3, s4, s5 string
        // var port, weight, timeout string
        // fmt.Sscanf(tline, "    server %s:%d weight=%d max_fails=2 fail_timeout=%ds", &ip, &port, &weight, &timeout) 
        fmt.Sscanf(line, "%s %s %s %s %s", &s1, &s2, &s3, &s4, &s5)
        s6 := strings.Split(s2, ":")
        ip := s6[0]
        port, _ := strconv.Atoi(s6[1])
        s7 := strings.Split(s3, "=")
        weight, _ := strconv.Atoi(s7[1])
        s8 := strings.Split(s5, "=")
        timeout, _ := strconv.Atoi(strings.Trim(s8[1], "s"))
        rs := model.RealServer{ip, port, weight, timeout}
        res = append(res, rs)
    }

    return res
}

// prepend servers specified
func NewServer(conf *model.NgxConfWrapper) error {
    file := buildFile(conf)
    exist := util.Exist(file)

    switch exist {
    case true:  // prepend server
        return insertServer(conf)        
    case false:  // new app and backend servers
        return newApp(conf)
    }
    return nil
}

func insertServer(conf *model.NgxConfWrapper) error {
    fileMutex.Lock()
    defer fileMutex.Unlock()

    indexs := findIndexs(conf)

    if len(indexs) == 0 {
        return nil
    }

    // build servers info and byte buffer
    tmpl, _ := template.New("server").Parse(SERVER)    

    buf := new(bytes.Buffer)
    for _, index := range indexs {
        buf.Write([]byte("    "))
        rs := conf.RealServers[index]
        tmpl.Execute(buf, rs)
        buf.Write([]byte("\n"))
    }

    // append the buffer to "upstream" line
    prependServers(conf, buf)
    
    return nil
}

func prependServers(conf *model.NgxConfWrapper, buff *bytes.Buffer) error {
    file := buildFile(conf)

    fil, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return nil 
    }
    defer fil.Close()

    match := fmt.Sprintf("backend_%s", conf.Domain)
    filebuff := new(bytes.Buffer)
    flag := false

    scanner := bufio.NewScanner(fil)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(line, "upstream") && strings.Contains(line, match) {
            flag = true
        }
        filebuff.Write([]byte(line))
        filebuff.Write([]byte{'\n'})
        if flag {
            filebuff.Write(buff.Bytes())
        }
        flag = false
    }

    fil2, err := os.OpenFile(file + ".tmp", os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return nil
    }
    fil2.Write(filebuff.Bytes())
    fil2.Sync()
    fil2.Close()

    os.Rename(file, file + ".bak")
    os.Rename(file + ".tmp", file)

    return nil
}

// index slice needs to be inserted
func findIndexs(conf *model.NgxConfWrapper) []int {
    // servers want to be inserted
    servers := genServers(conf) 

    file := buildFile(conf)

    fil, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return nil 
    }
    defer fil.Close()

    cons := make([]int, 0)

    scanner := bufio.NewScanner(fil)
    for scanner.Scan() {
        line := scanner.Text()
        for j, server := range servers {
            if strings.Contains(line, server) {
                cons = append(cons, j)
            }
        }
    }

    fmt.Println(cons)

    allLen := len(servers)
    res := make([]int, 0)
    for j := 0; j < allLen; j ++ {
        flag := true 
        for _, index := range cons {
            if index == j {
                flag = false
                break
            }
        }
        if flag {
            res = append(res, j)
        }
        flag = true
    }

    return res
}

// gen server slice like
// ["1.1.1.1:8001", "1.1.1.2:8002"]
func genServers(conf *model.NgxConfWrapper) []string {
    serversLen := len(conf.RealServers)
    servers := make([]string, serversLen)
    for j, server := range conf.RealServers {
        servers[j] = fmt.Sprintf("%s:%d", server.Ip, server.Port)
    }
    return servers
}

func DeleteServer(conf *model.NgxConfWrapper) (int, error) {
    file := buildFile(conf)
    exist := util.Exist(file)
    if !exist {
        return 0, errors.New("no such app domain")
    }

    // servers to deleted
    servers := genServers(conf)

    fileMutex.Lock()
    defer fileMutex.Unlock()

    fil, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return 0, errors.New("open file failed")
    }
    defer fil.Close()

    buffer.Reset()

    scanner := bufio.NewScanner(fil)
    flag := false
    for scanner.Scan() {
        line := scanner.Text()
        for _, server := range servers {
            if strings.Contains(line, server) {
                flag = true
                break
            }
        }
        if !flag {
            // write other lines to buffer
            buffer.Write([]byte(line))        
            buffer.Write([]byte{'\n'})
        }
        flag = false
    }

    fil2, err := os.OpenFile(file + ".tmp", os.O_RDWR|os.O_CREATE, 0644)
    fil2.Write(buffer.Bytes())
    fil2.Sync()
    fil2.Close()

    os.Rename(file, file + ".bak")
    os.Rename(file + ".tmp", file)

    return len(servers), nil
}

func ConfTest(file string) bool {
    out, _ := exec.Command("nginx", "-t").CombinedOutput()
    if strings.Contains(string(out), "test is successful") {
        return true
    }
    return false 
}

func ConfReload(file string) bool {
    out, _ := exec.Command("nginx", "-s", "reload").Output()
    if len(out) == 0 {
        return true
    }
    return false
}

func NgxStop(file string) bool {
    err := exec.Command("nginx", "-s", "stop").Run()
    if err != nil {
        return false
    }
    return true
}

func newApp(conf *model.NgxConfWrapper) error {
    tmpl, _ := template.ParseFiles("template/nginx/domain.conf.temp")
    buf := new(bytes.Buffer)
    tmpl.Execute(buf, conf)
    
    file := buildFile(conf)

    fileMutex.Lock()
    defer fileMutex.Unlock()

    fil, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
    defer fil.Close()

    if err != nil {
        return err
    }
    fil.Write(buf.Bytes())
    fil.Sync()
    return nil
}

func buildFile(conf *model.NgxConfWrapper) string {
    filename := fmt.Sprintf("%s.conf", conf.Domain)
    if conf.ConfPath == "" {
        conf.ConfPath = NGX_CONF_PATH
    }
    file := path.Join(conf.ConfPath, "conf.d", filename)
    return file
}
