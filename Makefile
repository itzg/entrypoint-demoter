build:
	go build ./cmd/entrypoint-demoter

test:
	go test ./...
	./test.sh

snapshot:
	goreleaser release --snapshot --rm-dist
