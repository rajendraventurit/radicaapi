package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rajendraventurit/radicaapi/handlers"
	"github.com/rajendraventurit/radicaapi/lib/db"
	"github.com/rajendraventurit/radicaapi/lib/env"
	"github.com/rajendraventurit/radicaapi/lib/logger"
	"github.com/rajendraventurit/radicaapi/lib/routetable"
)

var rTable = routetable.RouteTable{}

func main() {

	// Data store
	ldb, err := db.Connect("")
	if err != nil {
		log.Fatal(err)
	}
	defer ldb.Close()
	SetLocalDB(ldb)
	// http(s) Server
	conf, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// routing
	ev := env.New(ldb)
	rTable = handlers.GetRoutes(ev)
	router := routetable.NewRouter()
	router.AddMiddleware(newPermMiddleware)
	router.AddMiddleware(newTokenHandler)
	router.AddMiddleware(newHeaderHandler)
	router.AddMiddleware(newLogHandler)
	router.SetRouteTable(rTable)

	adr := fmt.Sprintf("%v:%v", conf.Host, conf.Port)
	if conf.KeyFile == "" || conf.CertFile == "" {
		logger.Message(fmt.Sprintf("Starting... http server at %s", adr))
		logger.Fatal(http.ListenAndServe(adr, router.Handler()))
	}
	logger.Message(fmt.Sprintf("Starting TLS server at %s", adr))
	logger.Fatal(http.ListenAndServeTLS(adr, conf.CertFile, conf.KeyFile, router.Handler()))
}

const defConfPath = "/etc/radica/server.json"

type config struct {
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

func loadConfig() (*config, error) {
	f, err := os.Open(defConfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	conf := config{}
	err = json.NewDecoder(f).Decode(&conf)
	return &conf, err
}
