.PHONY: start-db check-fmt test test-ci

start-db:
	docker run -e POSTGRES_PASSWORD=dbmigrat -e POSTGRES_USER=dbmigrat -d -p 5432:5432 postgres:13.3

check-fmt:
	DIFF=$$(gofmt -d .);echo "$${DIFF}";test -z "$${DIFF}"

test:
	go test -covermode=set -failfast

test-ci:
	go test -coverprofile=coverage.out -covermode=set
