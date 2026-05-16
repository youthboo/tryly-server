package repository

import "database/sql"

func IsNotFoundError(err error) bool {
	return err == sql.ErrNoRows
}
