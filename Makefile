parser:
	go run cmd/parser/main.go

test: test_text

test_text:
	go test -v ./pkg/text

cov_html: cov
	go tool cover -html=cover.out

cov:
	go test -tags integration -coverprofile cover.out ./pkg/*
