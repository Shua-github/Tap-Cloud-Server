package utils

import (
	"errors"
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/types"
	"gorm.io/gorm"
)

func DbIsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func ParseDbError(w http.ResponseWriter, err error) {
	if DbIsNotFound(err) {
		WriteError(w, types.TCSError{HTTPCode: http.StatusInternalServerError, TCSCode: types.DbNotFound, Message: "data not found"})
	} else {
		WriteError(w, types.TCSError{HTTPCode: http.StatusInternalServerError, TCSCode: types.DbError, Message: "Db Error:" + err.Error()})
	}
}
