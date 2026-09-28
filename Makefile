.PHONY:  build linux dev
BASE_PATH ?= /
GO ?= go
build:
	VITE_BASE_PATH=$(BASE_PATH) npm --prefix web run build
	$(GO) build -tags production -trimpath -o bin/feetable ./cmd/feetable
linux:
	VITE_BASE_PATH=$(BASE_PATH) npm --prefix web run build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -tags production -trimpath -o bin/feetable-linux-amd64 ./cmd/feetable
dev:
	$(GO) run ./cmd/feetable serve --db var/dev.sqlite

.PHONY: release deploy portal rollback releases
release:
	bash deploy/release.sh build
deploy:
	bash deploy/release.sh deploy
portal:
	bash deploy/release.sh portal
rollback:
	bash deploy/release.sh rollback $(COMMIT)
releases:
	bash deploy/release.sh list
