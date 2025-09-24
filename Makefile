.PHONY: bin

bin:
	go build -o bin/ ./cmd/...

image:
	docker build -t redis-multi-tenant-proxy:latest .
