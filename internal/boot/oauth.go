package boot

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthConfig struct {
	Google struct {
		ClientID     string `mapstructure:"client_id"`
		ClientSecret string `mapstructure:"client_secret"`
		RedirectURL  string `mapstructure:"redirect_url"`
	} `mapstructure:"google"`
}

func (c *Config) GoogleOAuth() oauth2.Config {
	return oauth2.Config{
		ClientID:     c.Auth.OAuth.Google.ClientID,
		ClientSecret: c.Auth.OAuth.Google.ClientSecret,
		RedirectURL:  c.Auth.OAuth.Google.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}
