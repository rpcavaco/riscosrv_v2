package main

import (
	"net"
	"path/filepath"
	"time"

	"github.com/jackc/pgx"
	// logging
)

// var ADDR_WS string = ":8094"
var DBCONNCFG = "dbconn_config.json"
var DBSQLSTATMTS = "sqlstatements.yaml"
var DBINITSMTGRP = "initprepared"
var LOG_MAXSIZE = 1 // megabytes
var LOG_MAXBACKUPS = 5
var LOG_MAXAGE = 28 //days
//var SRVROOT = "C:\\www\\docsplat"

type appServerConfig struct {
	Generic struct {
		Addr_hs          string `json:"addr_hs"`
		Logpath          string `json:"logpath"`
		Shutddelay_secs  int    `json:"shutddelay_secs"`
		Timedserver_mins int    `json:"timedserver_mins"`
	} `json:"generic"`
}

type appServer struct {
	http_listener net.Listener
	appPath       string
	cfg           *appServerConfig

	//ws_listener net.Listener
	db_connpool *pgx.ConnPool
	//ws_upgrader websocket.Upgrader
}

func (s *appServer) run() error {

	var err error

	s.db_connpool, err = DoPoolConnect(filepath.Join(s.appPath, DBCONNCFG), filepath.Join(s.appPath, DBSQLSTATMTS), DBINITSMTGRP)
	if err != nil {
		LogCriticalf("database: FATAL error, no connection: %s", err)
	} else {

		s.http_listener, err = AsyncListenAndServe(s.cfg.Generic.Addr_hs, s.hsmux, time.Duration(s.cfg.Generic.Shutddelay_secs)*time.Second, "HTTP")
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

func timedserve(appPath string, cfg *appServerConfig) {
	as := &appServer{appPath: appPath, cfg: cfg}
	err := as.run()
	if err != nil {
		LogInfo("Timed server closing in error")
		as.stop()
	} else {
		LogInfof("Timed server starting, live for %d mins", as.cfg.Generic.Timedserver_mins)
		time.Sleep(time.Duration(as.cfg.Generic.Timedserver_mins) * time.Minute)
		LogInfo("a fechar timed server")
		as.stop()
	}
}
