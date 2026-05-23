package dbutil

import "database/sql"

type QueryRower interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}
