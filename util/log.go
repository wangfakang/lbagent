package util

import (
    "strings"
    "github.com/siddontang/go-log/log"
)

var logger *log.Logger

func Log(v ...interface{}) {
    logger.Info(strings.Repeat("v% ", len(v)), v...)
}

func Loge(v ...interface{}) {
    logger.Error(strings.Repeat("v% ", len(v)), v...)
}

func SetLogger(logg *log.Logger) {
    logger = logg
}
