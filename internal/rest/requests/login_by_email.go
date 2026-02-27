package requests

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/pkg/resources"
)

func LoginByEmail(r *http.Request) (req resources.LoginByEmail, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = newDecodeError("body", err)
		return
	}

	errs := validation.Errors{
		"data/type":       validation.Validate(req.Data.Type, validation.Required, validation.In("account_session")),
		"data/attributes": validation.Validate(req.Data.Attributes, validation.Required),
	}
	return req, errs.Filter()
}
