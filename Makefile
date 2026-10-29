VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: web build dev test lint clean

web:
	npm --prefix web ci
	npm --prefix web run build

build: web
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o peng-top .

# pair with `npm --prefix web run dev` for the UI, it proxies /api through to this
dev:
	go run . --log-level debug

test:
	go test ./...
	npm --prefix web test --if-present

lint:
	go vet ./...
	test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

clean:
	rm -rf peng-top dist web/dist/assets web/dist/index.html
