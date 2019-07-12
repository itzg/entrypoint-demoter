build:
	go build .

test:
	go test .
	./test.sh

release:
	curl -sL https://git.io/goreleaser | bash

snapshot:
	goreleaser release --snapshot --rm-dist