package utils

import (
	"errors"
	"net/http"

	"gorm.io/gorm"
)

func DbIsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func ParseDbError(w http.ResponseWriter, err error) {
	if DbIsNotFound(err) {
		WriteError(w, http.StatusNotFound, "Not Found")
	} else {
		WriteError(w, http.StatusInternalServerError, "DB Error:"+err.Error())
	}
}
