package main

import (
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"

	"github.com/ishkong/go_ServerStatus_Client/Config"
	"github.com/ishkong/go_ServerStatus_Client/Global"
	"github.com/ishkong/go_ServerStatus_Client/Log_hook"
)

var conf *Config.Config


func main() {
	conf = Config.Get()

	rotateOptions := []rotatelogs.Option{
		rotatelogs.WithRotationTime(time.Hour * 24),
	}

	if conf.LogAging > 0 {
		rotateOptions = append(rotateOptions, rotatelogs.WithMaxAge(time.Hour*24*time.Duration(conf.LogAging)))
	} else {
		rotateOptions = append(rotateOptions, rotatelogs.WithMaxAge(time.Hour*24*365*10))
	}

	if conf.LogForceNew {
		rotateOptions = append(rotateOptions, rotatelogs.ForceNewFile())
	}

	w, err := rotatelogs.New(path.Join("logs", "%Y-%m-%d.log"), rotateOptions...)
	if err != nil {
		log.Errorf("rotatelogs init err: %v", err)
		panic(err)
	}

	log.AddHook(Log_hook.NewLocalHook(w, Log_hook.LogFormat{}, Log_hook.GetLogLevel(conf.LogLevel)...))

	log.Info("初始化完成")

	go Global.RunWebSocketClient(conf)

	<-Global.SetupMainSignalHandler()
}
