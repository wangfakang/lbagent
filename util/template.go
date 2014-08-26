package util

import (
    "bytes"
    "text/template"
    "github.com/astutesparrow/lbagent/model"
)

var (
    UPSTREAM_SERVER = `server {{.Ip}}:{{.Port}} weight={{if .Weight}}{{.Weight}}{{else}}10{{end}} max_fails=2 fail_timeout={{if .FailTimeout}}{{.FailTimeout}}{{else}}30{{end}}s;`
)

func NewUpServer(rs *model.RealServer) []byte {
    tmpl, err := template.New("server").Parse(UPSTREAM_SERVER)
    if err != nil {
        panic(err)
    }
    buff := new(bytes.Buffer)
    err = tmpl.Execute(buff, rs)
    if err != nil {
        panic(err)
    }

    return buff.Bytes()
}
