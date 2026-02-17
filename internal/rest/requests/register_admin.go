package requests

import (
	"encoding/json"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/netbill/auth-svc/pkg/resources"
	"github.com/netbill/restkit/tokens"
)

func newDecodeError(what string, err error) error {
	return validation.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}

func RegistrationAdmin(r *http.Request) (req resources.RegistrationAdmin, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = validation.Errors{
			"body": fmt.Errorf("decode request body: %w", err),
		}
		return
	}

	errs := validation.Errors{
		"data/type":       validation.Validate(req.Data.Type, validation.Required, validation.In("registration_account_by_admin")),
		"data/attributes": validation.Validate(req.Data.Attributes, validation.Required),

		"data/attributes/email": validation.Validate(
			req.Data.Attributes.Email, validation.Required, validation.Length(5, 255), is.Email),

		"data/attributes/role": validation.Validate(
			req.Data.Attributes.Role, validation.Required, validation.In(tokens.GetAllSystemUserRoles())),
	}

	return req, errs.Filter()
}
