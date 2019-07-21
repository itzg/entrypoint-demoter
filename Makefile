build:
	go build ./cmd/entrypoint-demoter

test:
	go test ./...
	./test.sh

release:
	curl -sL https://git.io/goreleaser | bash

snapshot:
	goreleaser release --snapshot --rm-dist