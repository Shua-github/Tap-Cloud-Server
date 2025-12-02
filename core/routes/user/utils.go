package user

import (
	"errors"
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func GetSession(r *http.Request, db utils.Db) (*utils.BoundStruct[Session], error) {
	tk := utils.GetSessionToken(r)
	if tk == "" {
		return nil, errors.New("invalid request(tk)")
	}
	return utils.Bind(db.NewTable("session"), tk, new(Session)), nil
}
