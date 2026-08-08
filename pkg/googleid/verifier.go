// Package googleid verifies Google-issued OpenID Connect ID tokens locally,
// without a round trip to Google's tokeninfo endpoint, by checking the JWT
// signature against Google's published JWKS.
package googleid

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuerHTTPS = "https://accounts.google.com"
	issuerBare  = "accounts.google.com"

	keysTTL = time.Hour
)

// certsURL is a var (rather than a const) so tests can point it at a fake JWKS server.
var certsURL = "https://www.googleapis.com/oauth2/v3/certs"

type claims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// Verifier validates Google ID tokens for a single OAuth client ID.
type Verifier struct {
	audience   string
	httpClient *http.Client

	fetchMu sync.Mutex

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func New(audience string) *Verifier {
	return &Verifier{
		audience:   audience,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		keys:       make(map[string]*rsa.PublicKey),
	}
}

// Verify checks the token's signature, issuer, audience and expiry, and
// returns the verified, already-confirmed-owned email address it carries.
func (v *Verifier) Verify(ctx context.Context, idToken string) (string, error) {
	var c claims

	keyfunc := func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("id token is missing kid header")
		}
		return v.publicKey(ctx, kid)
	}

	_, err := jwt.ParseWithClaims(idToken, &c, keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("parse id token: %w", err)
	}

	if c.Issuer != issuerHTTPS && c.Issuer != issuerBare {
		return "", fmt.Errorf("unexpected issuer %q", c.Issuer)
	}

	if !c.EmailVerified {
		return "", errors.New("google account email is not verified")
	}

	if c.Email == "" {
		return "", errors.New("id token has no email claim")
	}

	return c.Email, nil
}

func (v *Verifier) publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if key, ok := v.cachedKey(kid); ok {
		return key, nil
	}

	v.fetchMu.Lock()
	defer v.fetchMu.Unlock()

	// Someone else may have refreshed the cache while we were waiting for the lock.
	if key, ok := v.cachedKey(kid); ok {
		return key, nil
	}

	if err := v.fetchKeys(ctx); err != nil {
		return nil, err
	}

	key, ok := v.cachedKey(kid)
	if !ok {
		return nil, fmt.Errorf("unknown google signing key %q", kid)
	}
	return key, nil
}

func (v *Verifier) cachedKey(kid string) (*rsa.PublicKey, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if time.Since(v.fetchedAt) >= keysTTL {
		return nil, false
	}
	key, ok := v.keys[kid]
	return key, ok
}

func (v *Verifier) fetchKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certsURL, nil)
	if err != nil {
		return fmt.Errorf("build google jwks request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch google jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch google jwks: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("decode google jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(body.Keys))
	for _, k := range body.Keys {
		if k.Kty != "RSA" {
			continue
		}

		pub, err := rsaPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return errors.New("google jwks response had no usable RSA keys")
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()

	return nil
}

func rsaPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}

	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, errors.New("zero exponent")
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}
