IMAGE_NAME = "redis-multi-tenant-proxy"
TAG = "latest"
TARFILE_NAME = $(shell echo $(IMAGE_NAME)-$(TAG).tar | sed 's/\//-/g')

.PHONY: bin image save-image

bin:
	go build -o bin/ ./cmd/...

image:
	if [ -f ./$(TARFILE_NAME) ]; then \
		echo "Removing old tar file..."; \
		rm ./$(TARFILE_NAME); \
	fi
	docker build -t $(IMAGE_NAME):$(TAG) .

save-image:
	if [ -f ./$(TARFILE_NAME) ]; then \
		echo "Removing old tar file..."; \
		rm ./$(TARFILE_NAME); \
	fi
	docker save -o $(TARFILE_NAME) $(IMAGE_NAME):$(TAG)

