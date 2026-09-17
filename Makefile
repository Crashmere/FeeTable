.PHONY: test build linux dev
BASE_PATH ?= /
GO ?= go
test:
	$(GO) test ./...
	npm --prefix web test
	npm --prefix web run typecheck
	node --test deploy/deploy-release.test.mjs
build:
	VITE_BASE_PATH=$(BASE_PATH) npm --prefix web run build
	$(GO) build -tags production -trimpath -o bin/feetable ./cmd/feetable
linux:
	VITE_BASE_PATH=$(BASE_PATH) npm --prefix web run build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -tags production -trimpath -o bin/feetable-linux-amd64 ./cmd/feetable
dev:
	$(GO) run ./cmd/feetable serve --db var/dev.sqlite
