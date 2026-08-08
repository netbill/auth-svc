package controller

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/restkit/pagi"
	"github.com/stretchr/testify/require"
)

// fakeQRSessions implements sessionCore. QRConnect only ever calls
// CreateQRToken, so every other method panics if it's ever reached.
type fakeQRSessions struct {
	token string
	err   error
}

func (f *fakeQRSessions) CreateQRToken(context.Context) (string, error) { return f.token, f.err }

func (f *fakeQRSessions) LoginByEmail(context.Context, string, string) (models.TokensPair, error) {
	panic("not used by this test")
}

func (f *fakeQRSessions) LoginByGoogle(context.Context, string) (models.TokensPair, error) {
	panic("not used by this test")
}

func (f *fakeQRSessions) Refresh(context.Context, string) (models.TokensPair, error) {
	panic("not used by this test")
}

func (f *fakeQRSessions) GetMySession(context.Context, models.AccountActor, uuid.UUID) (models.Session, error) {
	panic("not used by this test")
}

func (f *fakeQRSessions) GetMySessions(
	context.Context, models.AccountActor, ...session.ListSessionsOption,
) (pagi.Page[[]models.Session], error) {
	panic("not used by this test")
}

func (f *fakeQRSessions) ConfirmQRToken(context.Context, models.AccountActor, string) (models.TokensPair, error) {
	panic("not used by this test")
}

func (f *fakeQRSessions) PublishQRToken(context.Context, string, []byte) error {
	panic("not used by this test")
}

func (f *fakeQRSessions) Logout(context.Context, models.AccountActor) error {
	panic("not used by this test")
}

func (f *fakeQRSessions) DeleteMySession(context.Context, models.AccountActor, uuid.UUID) error {
	panic("not used by this test")
}

func (f *fakeQRSessions) DeleteMySessions(context.Context, models.AccountActor) error {
	panic("not used by this test")
}

type fakeQRBus struct {
	ch chan []byte
}

func (f *fakeQRBus) SubscribeQRToken(context.Context, string) (<-chan []byte, func()) {
	return f.ch, func() {}
}

func newQRTestServer(t *testing.T, c *SessionController) *httptest.Server {
	t.Helper()

	testLog := log.New("debug", "text", "test")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.QRConnect(w, r.WithContext(scope.CtxLog(r.Context(), testLog)))
	})

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func TestQRConnect_DeliversTokensOverSSE(t *testing.T) {
	const qrToken = "550e8400-e29b-41d4-a716-446655440000"

	sessions := &fakeQRSessions{token: qrToken}
	bus := &fakeQRBus{ch: make(chan []byte, 1)}

	srv := newQRTestServer(t, &SessionController{sessions: sessions, bus: bus})

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	reader := bufio.NewReader(resp.Body)

	event, data := readSSEFrame(t, reader)
	require.Equal(t, "qr_token", event)
	require.JSONEq(t, `{"data":{"type":"qr_token","attributes":{"qr_token":"`+qrToken+`"}}}`, data)

	bus.ch <- []byte(`{"access":"a","refresh":"r"}`)

	event, data = readSSEFrame(t, reader)
	require.Equal(t, "tokens", event)
	require.Equal(t, `{"access":"a","refresh":"r"}`, data)
}

func TestQRConnect_CreateTokenFails(t *testing.T) {
	sessions := &fakeQRSessions{err: context.DeadlineExceeded}
	bus := &fakeQRBus{ch: make(chan []byte, 1)}

	srv := newQRTestServer(t, &SessionController{sessions: sessions, bus: bus})

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Get(srv.URL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func readSSEFrame(t *testing.T, r *bufio.Reader) (event, data string) {
	t.Helper()

	for {
		line, err := r.ReadString('\n')
		require.NoError(t, err)
		line = strings.TrimRight(line, "\n")

		switch {
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case line == "":
			if event != "" {
				return event, data
			}
		}
	}
}
