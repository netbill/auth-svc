package requests

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/pkg/oapi"
	"github.com/netbill/restkit"
)

func UpdateProfile(r *http.Request) (req oapi.UpdateProfile, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = restkit.NewDecodeError("body", err)
		return
	}

	errs := validation.Errors{
		"data/type":       validation.Validate(req.Data.Type, validation.Required, validation.In("user_profile")),
		"data/attributes": validation.Validate(req.Data.Attributes, validation.Required),
	}

	return req, errs.Filter()
}

func UpdateUsername(r *http.Request) (req oapi.UpdateUsername, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = restkit.NewDecodeError("body", err)
		return
	}

	errs := validation.Errors{
		"data/type":                validation.Validate(req.Data.Type, validation.Required, validation.In("user_username")),
		"data/attributes/username": validation.Validate(req.Data.Attributes.Username, validation.Required),
	}

	return req, errs.Filter()
}

func DeleteUploadUserAvatar(r *http.Request) (req oapi.DeleteUploadUserAvatar, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = restkit.NewDecodeError("body", err)
		return
	}

	errs := validation.Errors{
		"data/type":       validation.Validate(req.Data.Type, validation.Required, validation.In("user_avatar")),
		"data/attributes": validation.Validate(req.Data.Attributes, validation.Required),
	}

	return req, errs.Filter()
}
