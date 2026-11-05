.PHONY: dev build server test vet fmt clean

build: build-client build-server

build-client:
	cd client && bun install && bun run build
	rm -rf server/internal/web/dist
	cp -r client/dist server/internal/web/dist

build-server:
	cd server && go build -o ../marquee ./cmd/marquee

dev:
	bun run dev

server:
	cd server && set -a && . ../.env.local && set +a && go run ./cmd/marquee

test:
	cd server && go test ./...

vet:
	cd server && go vet ./...

fmt:
	cd server && gofmt -l -w .

clean:
	rm -rf marquee client/dist server/internal/web/dist/assets
