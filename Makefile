OPENAPI_GENERATOR := java -jar ~/openapi-generator-cli.jar
API_SRC := ./docs/rest/api.yaml
API_BUNDLED := ./docs/rest/api-bundled.yaml
DOCS_OUTPUT_DIR := ./docs/rest/web
DOCS_INTERNAL_DIR := ./docs/rest/web/docs
RESOURCES_DIR := ./pkg/oapi

bundle-oapi:
	test -d $(RESOURCES_DIR) || mkdir -p $(RESOURCES_DIR)
	test -d $(dir $(API_SRC)) || mkdir -p $(dir $(API_SRC))
	test -d $(dir $(API_BUNDLED)) || mkdir -p $(dir $(API_BUNDLED))
	test -d $(DOCS_OUTPUT_DIR) || mkdir -p $(DOCS_OUTPUT_DIR)

	rm -rf $(DOCS_INTERNAL_DIR) && mkdir -p $(DOCS_INTERNAL_DIR)
	rm -rf $(RESOURCES_DIR) && mkdir -p $(RESOURCES_DIR)
	swagger-cli bundle $(API_SRC) --outfile $(API_BUNDLED) --type yaml

	$(OPENAPI_GENERATOR) generate \
		-i $(API_BUNDLED) -g go \
		-o $(DOCS_OUTPUT_DIR) \
		--additional-properties=packageName=oapi \
		--import-mappings uuid.UUID=github.com/google/uuid --type-mappings string+uuid=uuid.UUID

	mkdir -p $(RESOURCES_DIR)
	find $(DOCS_OUTPUT_DIR) -name '*.go' -exec mv {} $(RESOURCES_DIR)/ \;
	find $(RESOURCES_DIR) -type f -name "*_test.go" -delete

build:
	go build -o ./cmd/auth-svc/main ./cmd/auth-svc/main.go

migrate-up:
	go build -o ./cmd/auth-svc/main ./cmd/auth-svc/main.go
	set -a && . ./.env && set +a && ./cmd/auth-svc/main migrate up

migrate-down:
	go build -o ./cmd/auth-svc/main ./cmd/auth-svc/main.go
	set -a && . ./.env && set +a && ./cmd/auth-svc/main migrate down

run-server:
	go build -o ./cmd/auth-svc/main ./cmd/auth-svc/main.go
	set -a && . ./.env && set +a && ./cmd/auth-svc/main run service

docker-up:
	docker compose -f deployment/docker-compose.yml up -d --build

docker-down:
	docker compose -f deployment/docker-compose.yml down

.PHONY: proto
proto:
	mkdir -p pkg/pb
	protoc \
	  --proto_path=proto \
	  --go_out=pkg/pb \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=pkg/pb \
	  --go-grpc_opt=paths=source_relative \
	  proto/*.proto

.PHONY: proto-doc
proto-doc:
	mkdir -p docs/grpc
	protoc \
	  --proto_path=proto \
	  --doc_out=docs/grpc \
	  --doc_opt=html,index.html \
	  proto/common.proto proto/account.proto proto/session.proto proto/auth.proto