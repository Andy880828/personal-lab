BIN      := scout
PKG      := github.com/Andy880828/personal-lab/scout
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0-dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS  := -s -w -X $(PKG)/internal/version.Version=$(VERSION) -X $(PKG)/internal/version.Commit=$(COMMIT)
PLATFORMS := linux/amd64 linux/arm64 darwin/arm64 windows/amd64

.PHONY: build test vet cross install clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BIN) ./cmd/scout

test:
	go test -race ./...

vet:
	go vet ./...
	GOOS=darwin go vet ./...
	GOOS=windows go vet ./...

install:
	CGO_ENABLED=0 go install -trimpath -ldflags '$(LDFLAGS)' ./cmd/scout

cross:
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=$$( [ $$os = windows ] && echo .exe ); \
		echo "build $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags '$(LDFLAGS)' -o dist/$(BIN)-$$os-$$arch$$ext ./cmd/scout || exit 1; \
	done

clean:
	rm -rf bin dist
