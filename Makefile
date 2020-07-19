build:
	go build ./cmd/entrypoint-demoter

test:
	go test ./...

snapshot:
	goreleaser release --snapshot --rm-dist
