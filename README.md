# RISCO Feature Server

RISCO feature server provides web services serving vector features on a GeoJSON-like protocol format.

It was developed in Golang using the [valyala/fasthttp](https://github.com/valyala/fasthttp) FastHTTP library. This library provides extremely low latency, being capable of  resolving hundreds of thousands of requests per second.

This a part of [RISCO Web Map solution](https://github.com/rpcavaco/risco). 

It's intended to connect to a PostgreSQL / PostGIS backend. In order to be properly used by RISCO feature server, a PostGIS backend database must contain a 'risco_v2' metadata *schema*. Please find the empty structure for such a *schema* at this [location](https://github.com/rpcavaco/riscosrv_v2_pg).

## Installation

1. create a folder and 'cd' to it
2. clone this repository into that folder
3. using a terminal opened on this folder run ./build.sh (or build.bat if you are in Windows)
4. create a subfolder called 'log'

To run it just type ...

	./riscosrv_v2

... and hit &lt;enter&gt;

It's easy to find documentation to create a systemd service from any executable, as [this example](https://linuxhandbook.com/create-systemd-services/). 

> [!IMPORTANT]
> In order to use RISCO server, DON'T FORGET  to provide a valid [configuration](#configuration).

## Use as a Window Service

It's easy to use command line utility *sc.exe* to create a new entry in Windows Services panel. Just follow these steps:

1. edit the contents of file 'install_win_service.bat' according to your needs
2. run it

When finished, there will be a new entry in Windows Services, using whatever is written in the *displayname* parameter in the file.

## Configuration

Configuration files are JSON and are supposed to be placed in same folder as as **riscosrv_v2** executable ('riscosrv_v2.exe' in Windows).

### Generic config 

Generic config is kept on **config.json** file. Some sample contents:

	{
		"generic": {
			"addr_hs": ":8020",	
			"logpath": "log/log.txt",
			"shutddelay_secs": 5,
			"timedserver_mins": 20
		}
	}

Parameters

- addr_hs: listen address and TCP port (separeted by ':')
- logpath: path of log file (example logs to 'log' folder created in previous steps)
- shutddelay_secs: currently unused but required parameter, any numerical will do
- timedserver_mins: currently unused but required parameter, any numerical will do

## Database connection config

Currently RISCO server works only with PostgreSQL / PostGIS.

Database connection configuration is stored on a file called **dbconn_config.json**. Some sample contents:

	{
		"user": "risco_v2",
		"password": "xxxxyyy",
		"host": "localhost",
		"port": 5432,
		"database": "mydata",
		"maxConnections": 4
	}

Parameters:

- user: database user (usually the owner of config 'risco_v2' *schema*)
- password: base 64 encoded database user's password
- host: database host address (IP or DNS name)
- port: TCP port
- database: database name
- maxConnections: number of max concurrent connections

