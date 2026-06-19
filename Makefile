BINARY := system-health-monitor

.PHONY: build run test fmt clean

build:
	go build -o $(BINARY) .

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w main.go collectors/*.go

clean:
	rm -f $(BINARY)
