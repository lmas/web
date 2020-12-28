
COVER=.cover.out
COVER_HTML=.cover.html

.PHONY: test
test:
	go test -race -coverprofile="${COVER}" ./...

.PHONY: cover
cover:
	go tool cover -html="${COVER}" -o="${COVER_HTML}"

.PHONY: lint
lint:
	golint ./...

.PHONY: bench
bench:
	go test -race -cover -bench=.

.PHONY: clean
clean:
	go clean
	rm "${COVER}"
	rm "${COVER_HTML}"
