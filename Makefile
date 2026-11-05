.PHONY: gen test run

# OpenAPI Code generieren
gen:
	go generate ./...

# Tests mit Race Detector
test:
	go test -v -race ./...

# Lokalen Server starten
run: gen
	go run cmd/server/main.go