package dberrors

import "fmt"

type DBError struct {
	Message string
}

func (e *DBError) Error() string {
	return fmt.Sprintf("DB Error: %s", e.Message)
}

func NewDBError(message string) *DBError {
	return &DBError{
		Message: message,
	}
}

func IsDBError(err error) bool {
	_, ok := err.(*DBError)
	return ok
}
