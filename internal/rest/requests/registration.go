package requests

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/resources"
)

func Registration(r *http.Request) (req resources.Registration, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = newDecodeError("body", err)
		return
	}

	errs := validation.Errors{
		"data/type":       validation.Validate(req.Data.Type, validation.Required, validation.In("registration_account")),
		"data/attributes": validation.Validate(req.Data.Attributes, validation.Required),
		"data/attributes/email": validation.Validate(
			req.Data.Attributes.Email, validation.Required, validation.Length(5, 255),
		),
	}

	return req, errs.Filter()
}
