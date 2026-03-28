BINARY  := netmon
CMD     := ./cmd/netmon
IMAGE   := netmon

.PHONY: build ui run clean vet docker docker-run

ui:
	cd web && npm install && npm run build

build: ui
	CGO_ENABLED=0 go build -o $(BINARY) $(CMD)

run: ui
	go run $(CMD)

clean:
	rm -f $(BINARY) netmon.db netmon.db-shm netmon.db-wal
	rm -rf web/dist web/node_modules

vet:
	go vet ./...

docker:
	docker build -f infra/Dockerfile -t $(IMAGE) .

docker-run: docker
	docker run --rm -p 8080:8080 --network host -v netmon-data:/data $(IMAGE)
