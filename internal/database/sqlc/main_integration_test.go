//go:build database
// +build database

package database

import (
	"flag"
	"log"
	"testing"
)

var testStore Store

func TestMain(m *testing.M) {
	var dsn string
	var err error

	flag.StringVar(&dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.Parse()

	testStore, err = NewStore(dsn)
	if err != nil {
		log.Fatalln("cannot connect to test database:", err)
	}
	m.Run()
}
