go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go-web % curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.63.4
