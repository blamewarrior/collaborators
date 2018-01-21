package blamewarrior

import "database/sql"

type SQLRunner interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	Prepare(string) (*sql.Stmt, error)
	Exec(string, ...interface{}) (sql.Result, error)
}

type rowsScanner interface {
	Columns() ([]string, error)
	Next() bool
	Close() error
	Err() error
	scanner
}

type scanner interface {
	Scan(dest ...interface{}) error
}
