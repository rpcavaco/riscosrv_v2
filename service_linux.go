package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"encoding/json"

	"gopkg.in/natefinch/lumberjack.v2" // logging
)

/*
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
*/

func getFireSignalsChannel() chan os.Signal {

	c := make(chan os.Signal, 1)
	signal.Notify(c, // https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		syscall.SIGHUP,  // "terminal is disconnected"
	)
	return c

}

func serviceMain(appPath string, cfg *appServerConfig) {

	exitChan := getFireSignalsChannel()

	as := &appServer{appPath: appPath, cfg: cfg}

	err := as.run()
	if err != nil {
		LogCriticalf("FATAL error in LINUX serviceMain: %s", err)
		as.stop()
	} else {
		LogInfof("Service running ...")
		<-exitChan
		LogInfof("Service terminated")
		as.stop()
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

	raw, err = os.ReadFile(filepath.Join(exPath, "config.json"))
	if err != nil {

		log.SetOutput(&lumberjack.Logger{
			Filename:   filepath.Join(exPath, "paniclogger.txt"),
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
				Filename:   filepath.Join(exPath, "paniclogger.txt"),
				MaxSize:    LOG_MAXSIZE,
				MaxBackups: LOG_MAXBACKUPS,
				MaxAge:     LOG_MAXAGE,
			})
			LogCriticalf("Startup, unmarshal error, config file %s: %v \n", filepath.Base("config.json"), err.Error())

		} else {

			log.SetOutput(&lumberjack.Logger{
				Filename:   filepath.Join(exPath, cfg.Generic.Logpath),
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
