package log

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/netbill/ape"
	"github.com/netbill/eventbox"
	"github.com/netbill/logium"
)

const (
	ServiceField   = "service"
	OperationField = "operation"
	ComponentField = "component"

	AccountIDField        = "account_id"
	AccountSessionIDField = "account_session_id"

	HTTPMethodField = "http_method"
	HTTPPathField   = "http_path"

	EventIDField       = "event_id"
	EventTypeField     = "event_type"
	EventTopicField    = "event_topic"
	EventVersionField  = "event_version"
	EventProducerField = "event_producer"
	EventAttemptField  = "event_attempt"
)

type Logger struct {
	base *slog.Logger
}

func New(level string, format string, serviceName string) *Logger {
	lvl := parseLevel(level)

	var handler slog.Handler

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		})

	default:
		handler = logium.NewAlignedTextHandler(os.Stdout, logium.AlignedTextOptions{
			Level:      lvl,
			TimeFormat: "2006-01-02 15:04:05",
			MsgWidth:   55,
			Colors:     true,
		})
	}

	base := slog.New(handler).
		With(slog.String(ServiceField, serviceName))

	return &Logger{base: base}
}

func (l *Logger) With(args ...any) logium.Logger {
	return &Logger{base: l.base.With(args...)}
}

func (l *Logger) WithFields(fields map[string]any) logium.Logger {
	if len(fields) == 0 {
		return l
	}
	args := make([]any, 0, len(fields))
	for k, v := range fields {
		args = append(args, slog.Any(k, v))
	}
	return &Logger{base: l.base.With(args...)}
}

func (l *Logger) WithField(key string, value any) logium.Logger {
	return &Logger{base: l.base.With(slog.Any(key, value))}
}

func (l *Logger) WithError(err error) logium.Logger {
	if err == nil {
		return l
	}

	var ae *ape.Error
	if errors.As(err, &ae) {
		return &Logger{base: l.base.With(
			slog.String("error", ae.Error()),
			slog.Any("ape", ae),
		)}
	}

	return &Logger{base: l.base.With(slog.String("error", err.Error()))}
}

func (l *Logger) WithOperation(operation string) logium.Logger {
	return &Logger{base: l.base.With(slog.String(OperationField, operation))}
}

func (l *Logger) WithComponent(component string) logium.Logger {
	return &Logger{base: l.base.With(slog.String(ComponentField, component))}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.base.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.base.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.base.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.base.Error(msg, args...)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.base.DebugContext(ctx, msg, args...)
}
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.base.InfoContext(ctx, msg, args...)
}
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.base.WarnContext(ctx, msg, args...)
}
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.base.ErrorContext(ctx, msg, args...)
}

func (l *Logger) WithRequest(r *http.Request) *Logger {
	if r == nil {
		return l
	}
	return &Logger{base: l.base.With(
		slog.String(HTTPMethodField, r.Method),
		slog.String(HTTPPathField, r.URL.Path),
	)}
}

func (l *Logger) WithAccountAuthClaims(auth interface {
	GetAccountID() uuid.UUID
	GetSessionID() uuid.UUID
}) *Logger {
	if auth == nil {
		return l
	}
	return &Logger{base: l.base.With(
		slog.String(AccountIDField, auth.GetAccountID().String()),
		slog.String(AccountSessionIDField, auth.GetSessionID().String()),
	)}
}

func (l *Logger) WithInboxEvent(ev eventbox.InboxEvent) *Logger {
	return &Logger{base: l.base.With(
		slog.String(EventIDField, ev.EventID.String()),
		slog.String(EventTopicField, ev.Topic),
		slog.String(EventTypeField, ev.Type),
		slog.Int(EventVersionField, int(ev.Version)),
		slog.String(EventProducerField, ev.Producer),
		slog.Int(EventAttemptField, int(ev.Attempts)),
	)}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
