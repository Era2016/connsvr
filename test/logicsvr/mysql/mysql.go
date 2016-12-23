package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/simplejia/connsvr/test/logicsvr/conf"

	_ "github.com/go-sql-driver/mysql"
	"github.com/simplejia/utils"
)

type Conf struct {
	ConnMaxLifetime string
	MaxIdleConns    int
	MaxOpenConns    int
	Dsn             string
}

var (
	DBS  map[string]*sql.DB = map[string]*sql.DB{}
	Envs map[string]*Conf
	Env  string
	C    *Conf
)

func parseDBFile(path string) {
	fcontent, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	fcontent = utils.RemoveAnnotation(fcontent)
	if err := json.Unmarshal(fcontent, &Envs); err != nil {
		panic(err)
	}

	Env = conf.Env
	C = Envs[Env]
	if C == nil {
		fmt.Println("env not right:", Env)
		os.Exit(-1)
	}

	db, err := sql.Open("mysql", C.Dsn)
	if err != nil {
		panic(err)
	}
	dur, _ := time.ParseDuration(C.ConnMaxLifetime)
	db.SetConnMaxLifetime(dur)
	db.SetMaxIdleConns(C.MaxIdleConns)
	db.SetMaxOpenConns(C.MaxOpenConns)

	key := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	DBS[key] = db
}

func init() {
	dir := "mysql"
	for i := 0; i < 3; i++ {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
		dir = filepath.Join("..", dir)
	}
	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) (reterr error) {
			if err != nil {
				reterr = err
				return
			}
			if info.IsDir() {
				return
			}
			if filepath.Ext(path) != ".json" {
				return
			}

			parseDBFile(path)
			return
		},
	)
	if err != nil {
		panic(err)
	}
}
