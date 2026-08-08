package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/responses"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
)

const operationGetMyUserAvatarUploadMediaLink = "get_my_user_avatar_upload_media_link"

func (c *UserController) CreateUploadMediaLink(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyUserAvatarUploadMediaLink)

	u, media, err := c.users.CreateUploadMediaLinks(r.Context(), scope.UserActor(r))
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		log.Info("user does not exist")
		render.ResponseError(w, problems.Unauthorized())
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, responses.UploadUserMediaLinks(r, u, media))
	}
}

const operationDeleteMyUserUploadAvatar = "delete_my_user_upload_avatar"

func (c *UserController) DeleteUploadMedia(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMyUserUploadAvatar)

	req, err := requests.DeleteUploadUserAvatar(r)
	if err != nil {
		log.WithError(err).Info("invalid delete upload user avatar request")
		render.ResponseError(w, problems.BadRequest(err)...)

		return
	}

	log = log.With("target_avatar_id", req.Data.Id)

	err = c.users.DeleteUploadMedia(
		r.Context(),
		scope.UserActor(r),
		user.DeleteUploadMediaParams{
			Avatar: req.Data.Attributes.AvatarKey,
		},
	)
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		log.WithError(err).Warn("user does not exist")
		render.ResponseError(w, problems.Unauthorized())

	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		render.Response(w, http.StatusOK, nil)
	}
}
