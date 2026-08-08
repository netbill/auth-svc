package rest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registerRequest — тело запроса регистрации в формате JSON:API
type registerRequest struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"attributes"`
	} `json:"data"`
}

// loginRequest — тело запроса логина по email
type loginRequest struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"attributes"`
	} `json:"data"`
}

// loginResponse — ответ логина с токенами доступа
type loginResponse struct {
	Data struct {
		Attributes struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"attributes"`
	} `json:"data"`
}

const testPassword = "!Test1234"

// TestRegistrationAndLogin последовательно регистрирует N аккаунтов и логинит каждый.
// N задаётся через load.accounts в test_config.yaml.
func TestRegistrationAndLogin(t *testing.T) {
	n := testCfg.Load.Accounts
	base := testCfg.REST.BaseURL

	for i := range n {
		email := fmt.Sprintf("test_%s@example.com", uuid.New().String()[:8])

		t.Run(fmt.Sprintf("account_%d", i+1), func(t *testing.T) {
			t.Parallel()

			// — регистрация —
			register(t, base, email)

			// — логин —
			tokens := login(t, base, email)
			assert.NotEmpty(t, tokens.Data.Attributes.AccessToken)
			assert.NotEmpty(t, tokens.Data.Attributes.RefreshToken)
		})
	}
}

func register(t *testing.T, base, email string) {
	t.Helper()

	var body registerRequest
	body.Data.Type = "account"
	body.Data.Attributes.Email = email
	body.Data.Attributes.Password = testPassword

	resp := do(t, http.MethodPost, base+"/auth-svc/v1/registration", body)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode,
		"registration failed for email=%s", email)
}

func login(t *testing.T, base, email string) loginResponse {
	t.Helper()

	var body loginRequest
	body.Data.Type = "account_session"
	body.Data.Attributes.Email = email
	body.Data.Attributes.Password = testPassword

	resp := do(t, http.MethodPost, base+"/auth-svc/v1/login/email", body)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"login failed for email=%s", email)

	var result loginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	return result
}

func do(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testClient.Do(req)
	require.NoError(t, err)

	return resp
}
