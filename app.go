package main

import (
	//"os"
	"encoding/json"
	"fmt"
	"strconv"

	//"regexp"
	//"errors"
	//"io/ioutil"

	//"github.com/satori/go.uuid"

	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"

	//"github.com/smallfish/simpleyaml"
	//"github.com/fasthttp-contrib/websocket"

	"golang.org/x/text/encoding/charmap"
)

func EncodeWindows1252(inp []byte) []byte {
	enc := charmap.ISO8859_1.NewEncoder()
	out, _ := enc.Bytes(inp)
	return out
}

/*
func validFileServerExtension(path string) bool {

	var ret bool = false

	if path == "/" {
		return true
	}

	lowerpath := strings.ToLower(path)

	htmpatt := regexp.MustCompile("\\.(htm[l]?|json)$")
	imgpatt := regexp.MustCompile("\\.(jp[e]?g|png|gif|tif[f]?)$")
	webpatt := regexp.MustCompile("\\.(svg|js|css|ttf)$")
	pltxtpatt := regexp.MustCompile("\\.(txt|md|mkd|csv)$")

	switch {
		case htmpatt.MatchString(lowerpath):
			ret = true
		case imgpatt.MatchString(lowerpath):
			ret = true
		case webpatt.MatchString(lowerpath):
			ret = true
		case pltxtpatt.MatchString(lowerpath):
			ret = true
	}

	return ret

}


// HTTP Server

var fs *fasthttp.FS = &fasthttp.FS{
		Root:               SRVROOT,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: false,
		Compress:           false,
		AcceptByteRange:    false,
	}



var fsHandler func(hsctx *fasthttp.RequestCtx) = fs.NewRequestHandler()
*/

func (s *appServer) featsHandler(hsctx *fasthttp.RequestCtx) {

	var vcnt, chunks, chunk int64
	var ferr error
	var outj string

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		sreqid := hsctx.QueryArgs().Peek("reqid")
		if len(sreqid) < 1 {
			LogErrorf("featsHandler parse params error, no reqid")
			hsctx.Error("featsHandler parse params error, no reqid", fasthttp.StatusInternalServerError)
			return
		}

		svertxcnt := hsctx.QueryArgs().Peek("vertxcnt")
		schunks := hsctx.QueryArgs().Peek("chunks")
		lname := hsctx.QueryArgs().Peek("lname")

		vcnt, ferr = strconv.ParseInt(string(svertxcnt), 10, 64)
		if ferr == nil {
			chunks, ferr = strconv.ParseInt(string(schunks), 10, 64)
		}
		if ferr == nil {
			schunk := hsctx.QueryArgs().Peek("chunk")
			if len(schunk) < 1 {
				chunk = 1
			} else {
				chunk, ferr = strconv.ParseInt(string(schunk), 10, 64)
			}
		}
		if ferr != nil {
			LogErrorf("featsHandler parse params error %s", ferr.Error())
			hsctx.Error("featsHandler parse params error", fasthttp.StatusInternalServerError)
		} else {
			if len(lname) < 1 {
				LogErrorf("featsHandler parse params error, no layer name")
				hsctx.Error("featsHandler parse params error, no layer name", fasthttp.StatusInternalServerError)
			} else {
				qryname := "initprepared.getfeats"
				LogTwitf("feats: %s %s %d %d %d", string(sreqid), lname, chunks, vcnt, chunk)

				row := s.db_connpool.QueryRow(qryname, sreqid, lname, chunks, vcnt, chunk)
				err := row.Scan(&outj)
				if err != nil {

					LogErrorf("featsHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
					hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)

				} else {

					fmt.Fprintf(hsctx, outj)
					hsctx.SetContentType("application/json; charset=utf8")

				}
			}
		}
	}
}

func (s *appServer) statsHandler(hsctx *fasthttp.RequestCtx) {

	var cenx, ceny, wid, hei, pixsz float64
	var ferr error
	var outj string

	scenx := hsctx.QueryArgs().Peek("cenx")
	sceny := hsctx.QueryArgs().Peek("ceny")
	swid := hsctx.QueryArgs().Peek("wid")
	shei := hsctx.QueryArgs().Peek("hei")
	spixsz := hsctx.QueryArgs().Peek("pixsz")

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		cenx, ferr = strconv.ParseFloat(string(scenx), 64)
		if ferr == nil {
			ceny, ferr = strconv.ParseFloat(string(sceny), 64)
		}
		if ferr == nil {
			wid, ferr = strconv.ParseFloat(string(swid), 64)
		}
		if ferr == nil {
			hei, ferr = strconv.ParseFloat(string(shei), 64)
		}
		if ferr == nil {
			pixsz, ferr = strconv.ParseFloat(string(spixsz), 64)
		}
		if ferr != nil {
			LogErrorf("statsHandler parse params error %s", ferr.Error())
			hsctx.Error("statsHandler parse params error", fasthttp.StatusInternalServerError)
		} else {

			mapname := hsctx.QueryArgs().Peek("map")
			if len(mapname) < 1 {
				LogErrorf("statsHandler parse params error, no map name")
				hsctx.Error("statsHandler parse params error, no map name", fasthttp.StatusInternalServerError)
			} else {
				vizlayers := hsctx.QueryArgs().Peek("vizlrs")
				filter_lname := hsctx.QueryArgs().Peek("flname")
				filter_fname := hsctx.QueryArgs().Peek("ffname")
				filter_value := hsctx.QueryArgs().Peek("fval")

				qryname := "initprepared.fullchunkcalc"
				LogTwitf("stats - cx:%f cy:%f pixsz:%f w:%f h:%f map:%s vizlyrs:%s filt_lname:%s fname:%s value:%s", cenx, ceny, pixsz, wid, hei, mapname, vizlayers, filter_lname, filter_fname, filter_value)
				row := s.db_connpool.QueryRow(qryname, cenx, ceny, pixsz, wid, hei, mapname, vizlayers, filter_lname, filter_fname, filter_value)
				err := row.Scan(&outj)
				if err != nil {
					LogErrorf("statsHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
					hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)
				} else {
					fmt.Fprintf(hsctx, outj)
					hsctx.SetContentType("application/json; charset=utf8")
				}
			}
		}

	}
}

type gJSONSaveElem struct {
	Lname  string `json:"lname"`
	Gisid  string `json:"gisid"`
	Userid string `json:"userid"`
	Epsg   int    `json:"epsg"`
	Gjson  struct {
		Type     string `json:"type"`
		Geometry struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"gjson"`
}

func (s *appServer) geojsonSaveHandler(hsctx *fasthttp.RequestCtx) {

	var gj gJSONSaveElem
	var qryname, sgjson, gid string
	var err error
	var b []byte
	var tx *pgx.Tx

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		LogInfof("geojsonSaveHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &gj); err != nil {

			LogErrorf("geojsonSaveHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {
			qryname = "initprepared.gjsonsave"

			b, err = json.Marshal(gj.Gjson)
			if err != nil {

				LogErrorf("geojson marshaling error %s", err.Error())
				hsctx.Error("geojson marshaling error", fasthttp.StatusInternalServerError)

			} else {

				sgjson = string(b)
				LogTwitf("lname:%s  gisid:%s userid:%s gjson:%s", gj.Lname, gj.Gisid, gj.Userid, sgjson)

				// Abrir transac????o
				tx, err = s.db_connpool.Begin()
				if err != nil {

					LogErrorf("geojsonInsert open transaction error %s", err.Error())
					hsctx.Error("transaction begin error", fasthttp.StatusInternalServerError)

				} else {

					defer tx.Rollback()

					// inserir local
					row := s.db_connpool.QueryRow(qryname, gj.Lname, gj.Gisid, gj.Userid, sgjson)
					err = row.Scan(&gid)
					if err != nil {
						LogErrorf("geojsonSaveHandler error %s, stmt name: '%s'", err.Error(), qryname)
						hsctx.Error("db error", fasthttp.StatusInternalServerError)
					} else {
						// Fechar transac????o
						err = tx.Commit()
						if err != nil {
							LogErrorf("geojsonSaveHandler commit transaction error %s", err.Error())
							hsctx.Error("commit error error", fasthttp.StatusInternalServerError)
						} else {
							fmt.Fprintf(hsctx, gid)
							hsctx.SetContentType("text/plain; charset=utf8")
						}
					}

				}

			}

		}
	}
}

type doGetElem struct {
	Alias      string   `json:"alias"`
	Filtervals []string `json:"filtervals"`
	Pbuffer    float32  `json:"pbuffer"`
	Lang       string   `json:"lang"`
}

func (s *appServer) doGetHandler(hsctx *fasthttp.RequestCtx) {

	var ge doGetElem
	var qryname string //, sgjson, gid string
	var outj []byte
	var err error

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {
		LogInfof("doGetHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &ge); err != nil {

			LogErrorf("doGetHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {
			qryname = "initprepared.doget"

			row := s.db_connpool.QueryRow(qryname, ge.Alias, ge.Filtervals, ge.Pbuffer, ge.Lang)
			err := row.Scan(&outj)
			if err != nil {
				LogErrorf("doGetHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
				hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)
			} else {
				fmt.Fprintf(hsctx, string(outj))
				hsctx.SetContentType("application/json; charset=utf8")
			}
		}
	}
}

func (s *appServer) testRequestHandler(hsctx *fasthttp.RequestCtx) {

	fmt.Fprintf(hsctx, "Hello, world!\n\n")
	hsctx.SetContentType("text/plain; charset=utf8")
}

func (s *appServer) hsmux(hsctx *fasthttp.RequestCtx) {
	LogTwitf("acesso HTTP: %s", hsctx.Path())
	switch string(hsctx.Path()) {
	case "/x":
		s.testRequestHandler(hsctx)
	case "/doget":
		s.doGetHandler(hsctx)
	case "/stats":
		s.statsHandler(hsctx)
	case "/feats":
		s.featsHandler(hsctx)
	/*case "/gjsonsave":
		s.geojsonSaveHandler(hsctx, true)
	case "/gjsonsaveg":
		s.geojsonSaveHandler(hsctx, false)
	*/
	//
	default:
		/*if validFileServerExtension(string(hsctx.Path())) {
			fsHandler(hsctx)
		} else { */
		hsctx.Error("not found", fasthttp.StatusNotFound)
		LogWarningf("HTTP not found: %s", string(hsctx.Path()))
		//}
	}
}
