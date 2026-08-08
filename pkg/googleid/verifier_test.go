package googleid

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const testAudience = "test-client-id.apps.googleusercontent.com"

func startJWKSServer(t *testing.T, kid string, pub *rsa.PublicKey) *httptest.Server {
	t.Helper()

	type jwk struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	}

	eBytes := big.NewInt(int64(pub.E)).Bytes()

	body, err := json.Marshal(struct {
		Keys []jwk `json:"keys"`
	}{
		Keys: []jwk{{
			Kty: "RSA",
			Kid: kid,
			N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(eBytes),
		}},
	})
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	return srv
}

func signToken(t *testing.T, key *rsa.PrivateKey, kid string, c claims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	token.Header["kid"] = kid

	signed, err := token.SignedString(key)
	require.NoError(t, err)

	return signed
}

func baseClaims() claims {
	now := time.Now()
	return claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerHTTPS,
			Audience:  jwt.ClaimStrings{testAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Email:         "user@example.com",
		EmailVerified: true,
	}
}

func TestVerifier_Verify_Success(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	srv := startJWKSServer(t, "kid-1", &key.PublicKey)
	orig := certsURL
	certsURL = srv.URL
	t.Cleanup(func() { certsURL = orig })

	token := signToken(t, key, "kid-1", baseClaims())

	v := New(testAudience)
	email, err := v.Verify(t.Context(), token)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", email)
}

func TestVerifier_Verify_WrongAudience(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	srv := startJWKSServer(t, "kid-1", &key.PublicKey)
	orig := certsURL
	certsURL = srv.URL
	t.Cleanup(func() { certsURL = orig })

	c := baseClaims()
	c.Audience = jwt.ClaimStrings{"someone-else.apps.googleusercontent.com"}
	token := signToken(t, key, "kid-1", c)

	v := New(testAudience)
	_, err = v.Verify(t.Context(), token)
	require.Error(t, err)
}

func TestVerifier_Verify_WrongIssuer(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	srv := startJWKSServer(t, "kid-1", &key.PublicKey)
	orig := certsURL
	certsURL = srv.URL
	t.Cleanup(func() { certsURL = orig })

	c := baseClaims()
	c.Issuer = "https://evil.example.com"
	token := signToken(t, key, "kid-1", c)

	v := New(testAudience)
	_, err = v.Verify(t.Context(), token)
	require.Error(t, err)
}

func TestVerifier_Verify_EmailNotVerified(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	srv := startJWKSServer(t, "kid-1", &key.PublicKey)
	orig := certsURL
	certsURL = srv.URL
	t.Cleanup(func() { certsURL = orig })

	c := baseClaims()
	c.EmailVerified = false
	token := signToken(t, key, "kid-1", c)

	v := New(testAudience)
	_, err = v.Verify(t.Context(), token)
	require.Error(t, err)
}

func TestVerifier_Verify_Expired(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	srv := startJWKSServer(t, "kid-1", &key.PublicKey)
	orig := certsURL
	certsURL = srv.URL
	t.Cleanup(func() { certsURL = orig })

	c := baseClaims()
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	token := signToken(t, key, "kid-1", c)

	v := New(testAudience)
	_, err = v.Verify(t.Context(), token)
	require.Error(t, err)
}

func TestVerifier_Verify_UnknownKid(t *testing.T) {
	signingKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// JWKS only advertises otherKey, token is signed with signingKey under a kid
	// that isn't in the set.
	srv := startJWKSServer(t, "kid-known", &otherKey.PublicKey)
	orig := certsURL
	certsURL = srv.URL
	t.Cleanup(func() { certsURL = orig })

	token := signToken(t, signingKey, "kid-unknown", baseClaims())

	v := New(testAudience)
	_, err = v.Verify(t.Context(), token)
	require.Error(t, err)
}

func TestVerifier_Verify_ForgedSignatureSameKid(t *testing.T) {
	legit, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	forged, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// JWKS advertises the legit key under "kid-1", but the token is actually
	// signed with a different (forged) key while claiming the same kid.
	srv := startJWKSServer(t, "kid-1", &legit.PublicKey)
	orig := certsURL
	certsURL = srv.URL
	t.Cleanup(func() { certsURL = orig })

	token := signToken(t, forged, "kid-1", baseClaims())

	v := New(testAudience)
	_, err = v.Verify(t.Context(), token)
	require.Error(t, err)
}
