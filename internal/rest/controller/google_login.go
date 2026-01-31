package controller

import (
	"net/http"

	"golang.org/x/oauth2"
)

func (c *Controller) LoginByGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	url := c.google.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
