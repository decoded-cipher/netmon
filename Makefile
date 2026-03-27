BINARY  := netmon
CMD     := ./cmd/netmon
IMAGE   := netmon

.PHONY: build run clean vet docker docker-run

build:
	CGO_ENABLED=0 go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

clean:
	rm -f $(BINARY) netmon.db netmon.db-shm netmon.db-wal

vet:
	go vet ./...

docker:
	docker build -t $(IMAGE) .

docker-run: docker
	docker run --rm -p 8080:8080 --network host -v netmon-data:/data $(IMAGE)
