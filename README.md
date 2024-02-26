# RISCO Feature Server

RISCO feature server provides web services serving vector features on a GeoJSON-like protocol format.

It was developed in Golang using the [valyala/fasthttp](https://github.com/valyala/fasthttp) FastHTTP library. This library provides extremely low latency, being capable of  resolving hundreds of thousands of requests per second.

This a part of [RISCO Web Map solution](https://github.com/rpcavaco/risco). 

It's intended to connect to a PostgreSQL / PostGIS backend. In order to be properly used by RISCO feature server, a PostGIS backend database must contain a 'risco_v2' metadata *schema*. You can find The structure for this *schema* at this [location](https://github.com/rpcavaco/riscosrv_v2_pg)
