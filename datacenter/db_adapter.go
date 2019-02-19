package datacenter

import (
	"database/sql"
)

type DBInterface interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
	Close() error
}

type TestDB struct {
}

func (db *TestDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (db *TestDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}

func (db *TestDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (db *TestDB) Close() error {
	return nil
}

func (db *TestDB) Begin() (*sql.Tx, error) {
	return nil, nil
}
