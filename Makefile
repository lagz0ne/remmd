.PHONY: build test race lint clean health

build:
	go build -o bin/remmd-bin ./cmd/remmd/

test:
	go test -count=1 -parallel=4 ./...

race:
	CGO_ENABLED=1 go test -race -count=1 ./... || go test -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

health: build
	./bin/remmd-bin health
