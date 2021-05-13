package main 

import (
	"os"
	"log"
	"net"
	"time"
	"flag"
	"path/filepath"
	//"fmt"
	//"errors"
	"io/ioutil"
	"encoding/json"
	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2" // logging
	"github.com/jackc/pgx"
	//"github.com/fasthttp-contrib/websocket"
)

//var ADDR_WS string = ":8094"
var DBCONNCFG = "\\dbconn_config.json"
var DBSQLSTATMTS = "\\sqlstatements.yaml"
var DBINITSMTGRP = "initprepared"
var LOG_MAXSIZE =   1 // megabytes
var LOG_MAXBACKUPS = 5
var LOG_MAXAGE = 28 //days
//var SRVROOT = "C:\\www\\docsplat"


type appServerConfig struct {
	Generic struct {
		Addr_hs string `json:"addr_hs"`
		Logpath string `json:"logpath"`
		Shutddelay_secs int `json:"shutddelay_secs"`
		Timedserver_mins int `json:"timedserver_mins"`
	} `json:"generic"`
}

type appServer struct {
	http_listener net.Listener
	appPath string
	cfg *appServerConfig

	//ws_listener net.Listener
	db_connpool *pgx.ConnPool
	//ws_upgrader websocket.Upgrader
}


func (s *appServer) Start(svc service.Service) error {
	// Start should not block. Do the actual work async.
	LogInfo("start servico")
	go s.run()
	return nil
}

func (s *appServer) run() error {
	
	var err error
	
	s.db_connpool, err = DoPoolConnect(s.appPath + DBCONNCFG, s.appPath + DBSQLSTATMTS, DBINITSMTGRP)
	if err != nil {
		LogCriticalf("database: FATAL error, no connection: %s", err)
	} else {
		
		err, s.http_listener = AsyncListenAndServe(s.cfg.Generic.Addr_hs, s.hsmux, time.Duration(s.cfg.Generic.Shutddelay_secs) * time.Second, "HTTP")
		if err != nil {
			LogCriticalf("httpserver: FATAL error in AsyncListenAndServe: %s", err)
		}
	}
	
	/* REMOVER para ter WebSockets
	s.prepareWebsockets()

	err, s.ws_listener = AsyncListenAndServe(ADDR_WS, s.wsmux, time.Duration(SHUTDDELAY_SECS) * time.Second, "websocket")
	if err != nil {
		LogCriticalf("wsockserver: FATAL error in AsyncListenAndServe: %s", err)
	}
	* */
	
	return err
}

func (s *appServer) stop() {
	s.db_connpool.Close()
	//s.ws_listener.Close()
	s.http_listener.Close()
}

func (s *appServer) Stop(svc service.Service) error {
	LogInfo("stop servico")
	s.stop()
	return nil
}

func timedserve(appPath string, cfg *appServerConfig) {
	as := &appServer{ appPath: appPath, cfg: cfg }
	err := as.run()
	if err != nil {
		LogInfo("Timed server closing in error")
		as.stop()
	} else {
		time.Sleep(time.Duration(as.cfg.Generic.Timedserver_mins) * time.Minute)
		LogInfo("a fechar timed server")
		as.stop()
	}
}

func serviceMain(appPath string, cfg *appServerConfig) {
	
	svcConfig := &service.Config{
		Name:   "fast HTTP Service",
		/*DisplayName: "XXXXX",
		Description: "yYYYYYYYY",*/
	}

	prg := &appServer{ appPath: appPath, cfg: cfg}
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
	
	raw, err = ioutil.ReadFile(exPath + "\\config.json")
	if err != nil {

		log.SetOutput(&lumberjack.Logger{
			Filename:   exPath + "\\paniclogger.txt",
			MaxSize: LOG_MAXSIZE,
			MaxBackups: LOG_MAXBACKUPS,
			MaxAge: LOG_MAXAGE,
		})
		
		LogCriticalf("Startup: %s", err.Error())
		
	} else {
	
		cfg = &appServerConfig{}
		err = json.Unmarshal(raw, cfg)
		if err != nil {
			
			log.SetOutput(&lumberjack.Logger{
				Filename:   exPath + "\\paniclogger.txt",
				MaxSize: LOG_MAXSIZE,
				MaxBackups: LOG_MAXBACKUPS,
				MaxAge: LOG_MAXAGE,
			})
			LogCriticalf("Startup, unmarshal error, config file %s: %v \n", filepath.Base("config.json"), err.Error())
			
		} else {

			log.SetOutput(&lumberjack.Logger{
				Filename:   exPath + cfg.Generic.Logpath,
				MaxSize: LOG_MAXSIZE,
				MaxBackups: LOG_MAXBACKUPS,
				MaxAge: LOG_MAXAGE,
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
