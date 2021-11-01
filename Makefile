main:
	go run ./main.go
parser:
	go run cmd/parser/main.go

test: test_text test_lang

test_lang:
	go test -tags integration ./pkg/lang

test_text:
	go test -tags integration ./pkg/text

cov_html: cov
	go tool cover -html=cover.out

cov:
	go test -tags integration -coverprofile cover.out ./pkg/*
