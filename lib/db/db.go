package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/rajendraventurit/radicaapi/lib/logger"
)

const defConfPath = "/etc/radica/database.json"

type config struct {
	Host       string  `json:"host"`
	Name       string  `json:"name"`
	Password   string  `json:"password"`
	Port       int64   `json:"port"`
	User       string  `json:"user"`
	Version    float64 `json:"version"`
	Migrations string  `json:"migrations"`
}

// Connect will attempt to connect to a database
// if path is blank it will use the defConfPath
func Connect(path string) (*sqlx.DB, error) {
	if path == "" {
		path = defConfPath
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	conf := config{}
	err = json.NewDecoder(f).Decode(&conf)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Name)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	logger.Message(fmt.Sprintf("Connecting to data store %v", conf.Host))
	// TODO
	//err = Migrate(db, conf.Migrations, conf.Version)
	return db, err
}

// ConnectLocal will connect to a local db for testing
func ConnectLocal() (*sqlx.DB, error) {
	host := "127.0.0.1"
	port := 3306
	user := "root"
	password := ""
	name := "radica"
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true", user, password, host, port, name)
	return sqlx.Connect("mysql", dsn)
}

// ConnectTesting will connect to the testing db
func ConnectTesting() (*sqlx.DB, error) {
	host := "testdb.cylbputtjaa0.us-west-2.rds.amazonaws.com"
	port := 3306
	user := ""
	password := ""
	name := "radica"
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true", user, password, host, port, name)
	return sqlx.Connect("mysql", dsn)
}

// Storer is a combined interface for a db
type Storer interface {
	Queryer
	Execer
	Transactor
}

// Transactor is an interface that creates transactions
type Transactor interface {
	Beginx() (*sqlx.Tx, error)
}

// Queryer is an interface for queries
type Queryer interface {
	sqlx.Queryer // Query, Queryx, QueryRowx
	Get(dest interface{}, query string, args ...interface{}) error
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	Select(dest interface{}, query string, args ...interface{}) error
}

// Execer is an interface for execution
type Execer interface {
	sqlx.Execer // Exec
	NamedExec(query string, arg interface{}) (sql.Result, error)
}
