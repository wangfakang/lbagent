package main

import (
    "fmt"
    "net/http"
    "path"
    "os"
    "github.com/astutesparrow/lbagent/model"
    "github.com/astutesparrow/lbagent/nginx"
    "github.com/astutesparrow/lbagent/util"
    "github.com/ant0ine/go-json-rest/rest"
    "github.com/siddontang/go-log/log"
)

func main() {
    initLog()

    handler := rest.ResourceHandler{XPoweredBy: "jd-json"}

    handler.SetRoutes(
        &rest.Route{"Get", "/nginx/#domain", func(w rest.ResponseWriter, req *rest.Request){
            domain := req.PathParam("domain")
            if domain == "" {
                rest.Error(w, "invalid request parameter", http.StatusBadRequest)
                return
            }
            servers := nginx.Servers(domain, "")
            w.WriteJson(&model.App{
                "test.jd.com",
                servers,
            }) 
        }},
        &rest.Route{"POST", "/nginx/#domain", func(w rest.ResponseWriter, req *rest.Request) {
            domain := req.PathParam("domain")
            if domain == "" {
                rest.Error(w, "invalid request parameter", http.StatusBadRequest)
                return
            }

            app := model.App{}
            req.DecodeJsonPayload(&app)
            conf := model.NgxConfWrapper{app, "/etc/nginx"}
            fmt.Println(conf)
            if app.Domain == "" || len(app.RealServers) == 0 {
                rest.Error(w, "invalid request parameter", http.StatusBadRequest)
                return
            }
            nginx.NewServer(&conf)
        }},
        &rest.Route{"DELETE", "/nginx/#domain", func(w rest.ResponseWriter, req *rest.Request) {
            domain := req.PathParam("domain")
            err := nginx.Delete(domain, "")
            if err != nil {
                rest.Error(w, "delete app error", http.StatusInternalServerError)
                return
            }
            w.WriteHeader(http.StatusOK)
        }},
        &rest.Route{"POST", "/nginx/delser/#domain", func(w rest.ResponseWriter, req *rest.Request) {
            domain := req.PathParam("domain")
            if domain == "" {
                rest.Error(w, "invalid request parameter", http.StatusBadRequest)
                return
            }
            app := model.App{}
            req.DecodeJsonPayload(&app)
            conf := model.NgxConfWrapper{app, "/etc/nginx"}
            if app.Domain == "" || len(app.RealServers) == 0 {
                rest.Error(w, "invalid request parameter", http.StatusBadRequest)
                return
            }
            _, err:= nginx.DeleteServer(&conf)
            if err != nil {
                rest.Error(w, "delete app error", http.StatusInternalServerError)
                return
            }
            w.WriteHeader(http.StatusOK)
        }},
    )
    http.ListenAndServe(":8080", &handler)
}

func initLog() *log.Logger {
    filename := path.Join("/var/log", "lbagent", "log")
    os.Mkdir(path.Join("/var/log", "lbagent"), 0655)
    h, err := log.NewRotatingFileHandler(filename, 1048576000, 3)
    if err != nil {
        panic(err)
    }
    logger := log.NewDefault(h)
    util.SetLogger(logger)
    return logger
}
