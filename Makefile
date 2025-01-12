migrate/new:
	migrate create -seq -ext sql -dir db/migrations ${name}

# add force 1 to force the migration to run if dirty
migrate/up:
	@migrate -path db/migrations -database sqlite3://${DB_FILEPATH} up

migrate/down:
	@migrate -path db/migrations -database sqlite3://${DB_FILEPATH} down 1

test:
	go test -race -shuffle=on ./...

coverage:
	go test -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

fmt:
	gofmt -s -w .

check-env:
ifndef DB_FILEPATH
	$(error DB_FILEPATH is undefined)
endif

commit:
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) fmt
	git add .
	git commit -m "${m}"

dev/run: check-env
	rm ${DB_FILEPATH}
	$(MAKE) migrate/up
	./scripts/seeddb.sh ${DB_FILEPATH}
	go run ./examples
