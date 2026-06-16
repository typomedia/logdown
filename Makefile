build: tidy
	CGO_ENABLED=0 go build -ldflags "-s -w" .

run: tidy
	go install github.com/air-verse/air@latest
	air

loc:
	go install github.com/boyter/scc/v3@latest
	scc --exclude-dir vendor --exclude-dir public .

check:
	go install github.com/client9/misspell/cmd/misspell@latest
	misspell -error app
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --disable=staticcheck
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 15 app
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -quiet --severity high ./...

update:
	go get -u
	go mod tidy

tidy:
	go mod tidy