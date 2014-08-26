package model

type RealServer struct {
    Ip string `json:"ip"`
    Port int `json:"port"`
    Weight int `json:"weight"`
    FailTimeout int `json:"timeout"`
}

type App struct {
    Domain string `json:"domain"`
    RealServers []RealServer `json:"servers"`
}

type NgxConfWrapper struct {
    App
    ConfPath string
}

func (app *App)Validate() bool {
    return true
}
