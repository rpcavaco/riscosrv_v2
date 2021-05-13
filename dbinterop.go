package main 

import (
	"fmt"
	"regexp"
	"strings"
	"errors"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"github.com/jackc/pgx"
	b64 "encoding/base64"
	"github.com/smallfish/simpleyaml"
)

// limpa todos os espaços a mais, nos extremos e no meio da string
func cleanspace(instr string) string {
	re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	ret := re_inside_whtsp.ReplaceAllString(strings.TrimSpace(instr), " ")
	
	return ret
}

func mustPrepare(db *pgx.ConnPool, name string, query string) error {
	LogTwitf("prepare name:%s", name)
	_, err := db.Prepare(name, query)
	return err
}

type connPoolFunction func(conn *pgx.ConnPool, yamlfilename string, groupname string) error

func CycleStatmentGroup(conn *pgx.ConnPool, yamlfilename string, groupname string, cpf connPoolFunction) error {
	
	source, err := ioutil.ReadFile(yamlfilename)
	if err != nil {
		return err
	}

	yaml, err := simpleyaml.NewYaml(source)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	init_stmt_map := yaml.Get(groupname)
	
	if *init_stmt_map == (simpleyaml.Yaml{}) {
		return errors.New(fmt.Sprintf("missing group '%s' in YAML source",groupname))
	}
	
	var err1 error
	var stmt string

	if init_stmt_map.IsMap() {
		keys, err := init_stmt_map.GetMapKeys()
		if err != nil {
			return err
		}
		for _, k := range keys {
			stmt, err1 = init_stmt_map.Get(k).String()
			if err1 != nil {
				return err1
			}
			// prepStmtMap[k] = mustPrepare(conn, k, cleanspace(stmt))
			err1 = cpf(conn, fmt.Sprintf("%s.%s", groupname, k), cleanspace(stmt))
			if err1 != nil {
				return err1
			}
		}		
	}
	
	return nil
}

func PrepareStatmentGroup(conn *pgx.ConnPool, yamlfilename string, groupname string) error {
	
	LogTwitf("yamlfilename:%s", yamlfilename)
	
	err := CycleStatmentGroup(conn, yamlfilename, groupname, mustPrepare)
	return err
}


func DoPoolConnect(p_configpath string, statmnts_yfname string, init_prepare_groupname string) (*pgx.ConnPool, error) {

	raw, err := ioutil.ReadFile(p_configpath)
	if err != nil {
		LogCriticalf("DoPoolConnect: %s", err.Error())
		return nil, err
	}
	
	var dbcfg = pgx.ConnPoolConfig{}
	err = json.Unmarshal(raw, &dbcfg)
	if err != nil {
		LogCriticalf("DoPoolConnect, unmarshal error, config file %s: %v \n", filepath.Base(p_configpath), err.Error())
		return nil, err
	}
	
	// descodificação da pwd
	decpwd, _ := b64.StdEncoding.DecodeString(dbcfg.Password)
    dbcfg.Password = string(decpwd)

	
	//fmt.Printf("host :%s\n", dbcfg.Host)

	connPool, errconn := pgx.NewConnPool(dbcfg)
	if errconn != nil {
		LogCriticalf("DoPoolConnect, unable to establish connection: %v\n", errconn)
		return nil, errconn
	}
	
	var err_general error
	
	var testv int
	if err_general = connPool.QueryRow("select 1").Scan(&testv); err_general != nil {
		LogCriticalf("DoPoolConnect, invalid connection, ping error: %v\n", err_general)
		connPool.Close()
		return nil, err_general
	}
	
	if len(statmnts_yfname) > 0 && len(init_prepare_groupname) > 0 {
		err_general = PrepareStatmentGroup(connPool, statmnts_yfname, init_prepare_groupname)
		if err_general != nil {
			LogCriticalf("DoPoolConnect, initial statement prepare error: %v\n", err_general)
			connPool.Close()
			return nil, err_general
		}
	}

	return connPool, nil
}



/*

func main() {

	conn, err := DoPoolConnect("dbconn_config.json", "sqlstatements.yaml", "initprepared")
	if err != nil {
		fmt.Println("sem ligacao")
		return
	}
	defer conn.Close()
	
	// prepstatemtents := make(map[string]*pgx.PreparedStatement)
	
	var toponimo string
	
	err = conn.QueryRow("initprepared.toponimos").Scan(&toponimo)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	
	fmt.Printf("Topo: %s\n", toponimo)

}

*/

