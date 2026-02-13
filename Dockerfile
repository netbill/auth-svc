# ==============================
# 1) BUILD STAGE
# ==============================
ARG GO_VERSION=1.25.4
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /service
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o main \
    ./cmd/auth-svc

# ==============================
# 2) FINAL STAGE
# ==============================
FROM alpine:latest

WORKDIR /service

RUN apk add --no-cache ca-certificates

COPY --from=builder /service/main .
COPY config.yaml .

ENV KV_VIPER_FILE=/service/config.yaml

CMD ["./main", "run", "service"]

