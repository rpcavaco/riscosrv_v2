package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

/*

func EncodeWindows1252(inp []byte) []byte {
	enc := charmap.ISO8859_1.NewEncoder()
	out, _ := enc.Bytes(inp)
	return out
}
*/

func (s *appServer) addCORSHeaders(hsctx *fasthttp.RequestCtx) {
	hsctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	hsctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	hsctx.Response.Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
}

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

					fmt.Fprint(hsctx, outj)
					hsctx.SetContentType("application/json; charset=utf8")
					s.addCORSHeaders(hsctx)

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
				/* filter_lname := hsctx.QueryArgs().Peek("flname")
				filter_fname := hsctx.QueryArgs().Peek("ffname")
				filter_value := hsctx.QueryArgs().Peek("fval") */

				qryname := "initprepared.fullchunkcalc"
				LogTwitf("stats - cx:%f cy:%f pixsz:%f w:%f h:%f map:%s vizlyrs:%s", cenx, ceny, pixsz, wid, hei, mapname, vizlayers)
				row := s.db_connpool.QueryRow(qryname, cenx, ceny, pixsz, wid, hei, mapname, vizlayers)
				err := row.Scan(&outj)
				if err != nil {
					LogErrorf("statsHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
					hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)
				} else {
					fmt.Fprint(hsctx, outj)
					hsctx.SetContentType("application/json; charset=utf8")
					s.addCORSHeaders(hsctx)
				}
			}
		}

	}
}

type JSONSaveElem struct {
	Lname       string `json:"lname"`
	SessionId   string `json:"sessionid"`
	Mapname     string `json:"mapname"`
	Featholders []struct {
		Gisid string `json:"gisid,omitempty"`
		Feat  struct {
			Type     string `json:"type"`
			Geometry struct {
				Type        string    `json:"type"`
				Crs         int       `json:"crs"`
				Coordinates []float64 `json:"coordinates"`
			} `json:"geometry,omitempty"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		} `json:"feat,omitempty"`
	} `json:"featholders"`
}

func (s *appServer) saveHandler(hsctx *fasthttp.RequestCtx) {

	var jse JSONSaveElem
	var jsr map[string]interface{}
	var qryname, sjson string
	var err error
	var b []byte
	var tx *pgx.Tx

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		LogInfof("saveHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &jse); err != nil {

			LogErrorf("saveHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {

			rub := hsctx.Request.Header.Peek("Remote-User")
			ru := string(rub)

			qryname = "initprepared.save"

			b, err = json.Marshal(jse.Featholders)
			if err != nil {

				LogErrorf("json feature marshaling error %s", err.Error())
				hsctx.Error("json feature marshaling error", fasthttp.StatusInternalServerError)

			} else {

				sjson = string(b)
				LogTwitf("lname:%s session:%s map:%s feats.json:%s ru:%s", jse.Lname, jse.SessionId, jse.Mapname, sjson, ru)

				// Abrir transacção
				tx, err = s.db_connpool.Begin()
				if err != nil {

					LogErrorf("json insert open transaction error %s", err.Error())
					hsctx.Error("transaction begin error", fasthttp.StatusInternalServerError)

				} else {

					defer tx.Rollback()

					// inserir local
					row := s.db_connpool.QueryRow(qryname, jse.Lname, jse.SessionId, sjson, jse.Mapname, ru)
					err = row.Scan(&jsr)
					if err != nil {
						LogErrorf("saveHandler error %s, stmt name: '%s'", err.Error(), qryname)
						hsctx.Error("db error", fasthttp.StatusInternalServerError)
					} else {
						// Fechar transacção
						err = tx.Commit()
						if err != nil {
							LogErrorf("saveHandler commit transaction error %s", err.Error())
							hsctx.Error("commit error error", fasthttp.StatusInternalServerError)
						} else {

							b, err = json.Marshal(jsr)
							if err != nil {

								LogErrorf("json save response marshaling error %s", err.Error())
								hsctx.Error("json save response marshaling error", fasthttp.StatusInternalServerError)

							} else {

								fmt.Fprint(hsctx, string(b))
								hsctx.SetContentType("text/plain; charset=utf8")
								s.addCORSHeaders(hsctx)

							}

						}
					}

				}

			}

		}
	}
}

type doGetElem struct {
	Alias      string   `json:"alias"`
	Keyword    string   `json:"keyword"`
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

			row := s.db_connpool.QueryRow(qryname, ge.Alias, ge.Keyword, ge.Filtervals, ge.Pbuffer, ge.Lang)
			err := row.Scan(&outj)
			if err != nil {
				LogErrorf("doGetHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
				hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)
			} else {
				fmt.Fprint(hsctx, string(outj))
				hsctx.SetContentType("application/json; charset=utf8")
				s.addCORSHeaders(hsctx)
			}
		}
	}
}

type alphaStatsElem struct {
	Key     string `json:"key"`
	Options struct {
		Outsrid     int    `json:"outsrid,omitempty"`
		Clustersize int    `json:"clustersize,omitempty"`
		Col         string `json:"col,omitempty"`
	} `json:"options"`
}

func (s *appServer) alphaStatsHandler(hsctx *fasthttp.RequestCtx) {

	var ase alphaStatsElem
	var qryname string //, sgjson, gid string
	var outj []byte
	var err error

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		LogInfof("alphaStatsHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &ase); err != nil {

			LogErrorf("alphaStatsHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {

			qryname = "initprepared.astats"

			row := s.db_connpool.QueryRow(qryname, ase.Key, ase.Options)
			err := row.Scan(&outj)
			if err != nil {
				LogErrorf("alphaStatsHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
				hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)
			} else {
				fmt.Fprint(hsctx, string(outj))
				hsctx.SetContentType("application/json; charset=utf8")
				s.addCORSHeaders(hsctx)
			}
		}
	}
}

type InsertSelElem struct {
	Selname string `json:"selname"`
	Seldata struct {
		Desc  string `json:"desc"`
		Elems []struct {
			Gisid    string `json:"gisid"`
			Gislabel string `json:"gislabel"`
		} `json:"elems"`
	} `json:"seldata"`
}

func (s *appServer) insertselHandler(hsctx *fasthttp.RequestCtx) {

	var ise InsertSelElem
	var jsr map[string]interface{}
	var qryname, sjson string
	var err error
	var b []byte
	var tx *pgx.Tx

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		LogInfof("insertselHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &ise); err != nil {

			LogErrorf("saveHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {

			qryname = "initprepared.insertsel"

			b, err = json.Marshal(ise.Seldata)
			if err != nil {

				LogErrorf("json feature marshaling error %s", err.Error())
				hsctx.Error("json feature marshaling error", fasthttp.StatusInternalServerError)

			} else {

				sjson = string(b)
				LogTwitf("selname:%s payload:%s", ise.Selname, sjson)

				// Abrir transacção
				tx, err = s.db_connpool.Begin()
				if err != nil {

					LogErrorf("json insert open transaction error %s", err.Error())
					hsctx.Error("transaction begin error", fasthttp.StatusInternalServerError)

				} else {

					defer tx.Rollback()

					// inserir local
					row := s.db_connpool.QueryRow(qryname, ise.Selname, sjson)
					err = row.Scan(&jsr)
					if err != nil {
						LogErrorf("insertselHandler error %s, stmt name: '%s'", err.Error(), qryname)
						hsctx.Error("db error", fasthttp.StatusInternalServerError)
					} else {
						// Fechar transacção
						err = tx.Commit()
						if err != nil {
							LogErrorf("insertselHandler commit transaction error %s", err.Error())
							hsctx.Error("commit error error", fasthttp.StatusInternalServerError)
						} else {

							b, err = json.Marshal(jsr)
							if err != nil {

								LogErrorf("insertselHandler response marshaling error %s", err.Error())
								hsctx.Error("json save response marshaling error", fasthttp.StatusInternalServerError)

							} else {

								fmt.Fprint(hsctx, string(b))
								hsctx.SetContentType("text/plain; charset=utf8")
								s.addCORSHeaders(hsctx)

							}

						}
					}

				}

			}

		}
	}
}

type GetSelElem struct {
	Selname string `json:"selname"`
	Selcode string `json:"selcode"`
}

func (s *appServer) getselHandler(hsctx *fasthttp.RequestCtx) {

	var gse GetSelElem
	var jsr map[string]interface{}
	var qryname string
	var err error
	var b []byte
	var tx *pgx.Tx

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		LogInfof("insertselHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &gse); err != nil {

			LogErrorf("saveHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {

			qryname = "initprepared.getsel"

			LogTwitf("selname:%s payload:%s", gse.Selname, gse.Selcode)

			// Abrir transacção
			tx, err = s.db_connpool.Begin()
			if err != nil {

				LogErrorf("json insert open transaction error %s", err.Error())
				hsctx.Error("transaction begin error", fasthttp.StatusInternalServerError)

			} else {

				defer tx.Rollback()

				// inserir local
				row := s.db_connpool.QueryRow(qryname, gse.Selname, gse.Selcode)
				err = row.Scan(&jsr)
				if err != nil {
					LogErrorf("insertselHandler error %s, stmt name: '%s'", err.Error(), qryname)
					hsctx.Error("db error", fasthttp.StatusInternalServerError)
				} else {
					// Fechar transacção
					err = tx.Commit()
					if err != nil {
						LogErrorf("insertselHandler commit transaction error %s", err.Error())
						hsctx.Error("commit error error", fasthttp.StatusInternalServerError)
					} else {

						b, err = json.Marshal(jsr)
						if err != nil {

							LogErrorf("insertselHandler response marshaling error %s", err.Error())
							hsctx.Error("json save response marshaling error", fasthttp.StatusInternalServerError)

						} else {

							fmt.Fprint(hsctx, string(b))
							hsctx.SetContentType("text/plain; charset=utf8")
							s.addCORSHeaders(hsctx)

						}

					}
				}

			}

		}
	}
}

/*
type binnParamsElem struct {
	Key      string  `json:"key"`
	Geomtype string  `json:"geomtype"`
	Radius   float64 `json:"radius"`
}

func (s *appServer) binningHandler(hsctx *fasthttp.RequestCtx) {

	var bpe binnParamsElem
	var qryname string
	var outj []byte
	var err error

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		LogInfof("binningHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &bpe); err != nil {

			LogErrorf("binningHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {

			qryname = "initprepared.binning"

			row := s.db_connpool.QueryRow(qryname, bpe.Key, bpe.Geomtype, bpe.Radius)
			err := row.Scan(&outj)
			if err != nil {
				LogErrorf("binningHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
				hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)
			} else {
				fmt.Fprintf(hsctx, string(outj))
				hsctx.SetContentType("application/json; charset=utf8")
			}
		}
	}
}

*/

func (s *appServer) testRequestHandler(hsctx *fasthttp.RequestCtx) {

	fmt.Fprintf(hsctx, "Hello, world!\n\n")
	hsctx.SetContentType("text/plain; charset=utf8")
}

type locateElemElem struct {
	Mapname string `json:"mapname"`
	Lname   string `json:"lname"`
	Gisid   string `json:"gisid"`
}

func (s *appServer) locateElemHandler(hsctx *fasthttp.RequestCtx) {

	var lee locateElemElem
	var qryname string //, sgjson, gid string
	var outj []byte
	var err error

	if string(hsctx.Method()) == "OPTIONS" {

		fmt.Fprintf(hsctx, "ok")
		hsctx.SetContentType("text/plain; charset=utf8")

	} else {

		LogInfof("locateElemHandler, body:'%s'", hsctx.PostBody())

		if err = json.Unmarshal(hsctx.PostBody(), &lee); err != nil {

			LogErrorf("locateElemHandler generic unmarshal error: %s body:'%s'", err.Error(), hsctx.PostBody())
			hsctx.Error("unmarshal error", fasthttp.StatusInternalServerError)

		} else {

			qryname = "initprepared.locateelem"

			row := s.db_connpool.QueryRow(qryname, lee.Mapname, lee.Lname, lee.Gisid)
			err := row.Scan(&outj)
			if err != nil {
				LogErrorf("locateElemHandler dbquery return read error %s, stmt name: '%s'", err.Error(), qryname)
				hsctx.Error("dbquery return read error", fasthttp.StatusInternalServerError)
			} else {
				fmt.Fprint(hsctx, string(outj))
				hsctx.SetContentType("application/json; charset=utf8")
				s.addCORSHeaders(hsctx)
			}
		}
	}

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
	case "/astats":
		s.alphaStatsHandler(hsctx)
	case "/save":
		s.saveHandler(hsctx)
	case "/locateelem":
		s.locateElemHandler(hsctx)
	case "/insertsel":
		s.insertselHandler(hsctx)
	case "/getsel":
		s.getselHandler(hsctx)

	default:
		/*if validFileServerExtension(string(hsctx.Path())) {
			fsHandler(hsctx)
		} else { */
		hsctx.Error("not found", fasthttp.StatusNotFound)
		LogWarningf("HTTP not found: %s", string(hsctx.Path()))
		//}
	}
}
