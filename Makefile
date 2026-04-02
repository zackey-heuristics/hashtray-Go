VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
BINARY := hashtray-go

.PHONY: build test lint clean cross

build:
	CGO_ENABLED=0 go build -buildvcs=false $(LDFLAGS) -o $(BINARY) ./cmd/hashtray

test:
	CGO_ENABLED=0 go test -buildvcs=false ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BINARY) dist/

cross:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/hashtray-go-linux-amd64 ./cmd/hashtray
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/hashtray-go-linux-arm64 ./cmd/hashtray
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/hashtray-go-darwin-amd64 ./cmd/hashtray
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/hashtray-go-darwin-arm64 ./cmd/hashtray
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/hashtray-go-windows-amd64.exe ./cmd/hashtray
