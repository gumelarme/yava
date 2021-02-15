parser:
	go run cmd/parser/main.go

test: test_text
	go test -v ./pkg/*

test_text:
	go test -v ./pkg/text

cov_html: cov
	go tool cover -html=cover.out

cov:
	go test -coverprofile cover.out ./pkg/*
