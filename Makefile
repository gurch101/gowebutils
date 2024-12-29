migrate/new:
	migrate create -seq -ext sql -dir db/migrations ${name}

# add force 1 to force the migration to run if dirty
migrate/up:
	@migrate -path db/migrations -database ${MIGRATE_DSN} up

migrate/down:
	@migrate -path db/migrations -database ${MIGRATE_DSN} down 1

test:
	go test -v ./...

fmt:
	go fmt ./...

dev/run:
	rm -rf ${DEV_DB_FILEPATH}
	$(MAKE) migrate/up
	go run .
