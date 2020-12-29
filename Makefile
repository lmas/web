
COVER=.cover.out
COVER_HTML=.cover.html

.PHONY: test
test:
	go test -race -coverprofile="${COVER}" ./...

.PHONY: cover
cover:
	go tool cover -html="${COVER}" -o="${COVER_HTML}"

.PHONY: bench
bench:
	go test -race -cover -bench=.

.PHONY: lint
lint:
	golint ./...

.PHONY: secure
secure:
	gosec -quiet -fmt=golint ./...

.PHONY: clean
clean:
	go clean
	rm "${COVER}"
	rm "${COVER_HTML}"
