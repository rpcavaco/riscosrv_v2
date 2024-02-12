package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"encoding/json"

	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2" // logging
)

func (s *appServer) Start(svc service.Service) error {
	// Start should not block. Do the actual work async.
	LogInfo("start servico")
	go s.run()
	return nil
}

func (s *appServer) Stop(svc service.Service) error {
	LogInfo("stop servico")
	s.stop()
	return nil
}

func serviceMain(appPath string, cfg *appServerConfig) {

	svcConfig := &service.Config{
		Name: "fast HTTP Service",
		/*DisplayName: "XXXXX",
		Description: "yYYYYYYYY",*/
	}

	prg := &appServer{appPath: appPath, cfg: cfg}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		LogError(err.Error())
	}

	err = s.Run()
	if err != nil {
		LogCritical(err.Error())
	}
}

func main() {

	var noservice bool
	var err error
	var raw []byte
	var ex string
	var cfg *appServerConfig

	flag.BoolVar(&noservice, "noservice", false, "como processo autonomo")
	flag.Parse()

	ex, err = os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	raw, err = os.ReadFile(exPath + "\\config.json")
	if err != nil {

		log.SetOutput(&lumberjack.Logger{
			Filename:   exPath + "\\paniclogger.txt",
			MaxSize:    LOG_MAXSIZE,
			MaxBackups: LOG_MAXBACKUPS,
			MaxAge:     LOG_MAXAGE,
		})

		LogCriticalf("Startup: %s", err.Error())

	} else {

		cfg = &appServerConfig{}
		err = json.Unmarshal(raw, cfg)
		if err != nil {

			log.SetOutput(&lumberjack.Logger{
				Filename:   exPath + "\\paniclogger.txt",
				MaxSize:    LOG_MAXSIZE,
				MaxBackups: LOG_MAXBACKUPS,
				MaxAge:     LOG_MAXAGE,
			})
			LogCriticalf("Startup, unmarshal error, config file %s: %v \n", filepath.Base("config.json"), err.Error())

		} else {

			log.SetOutput(&lumberjack.Logger{
				Filename:   exPath + cfg.Generic.Logpath,
				MaxSize:    LOG_MAXSIZE,
				MaxBackups: LOG_MAXBACKUPS,
				MaxAge:     LOG_MAXAGE,
			})
		}
	}

	if err == nil {

		//noservice = true
		if noservice {
			timedserve(exPath, cfg)
		} else {
			serviceMain(exPath, cfg)
		}

	}
}
