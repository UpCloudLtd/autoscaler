
UPCLOUD_SDK_PACKAGE ?= github.com/upcloudltd/upcloud-go-api/v6
# TODO: change to v6.3.2 when released
UPCLOUD_SDK_VERSION ?= main 

IMAGE_TAG ?= ghcr.io/peknur/uks-ca:dev

build:
	BUILD_TAGS=upcloud make -C ../../ build-in-docker

image: build
	docker build -t $(IMAGE_TAG) -f ../../Dockerfile.amd64 ../../

vendor:
	sh vendor.sh $(UPCLOUD_SDK_PACKAGE) $(UPCLOUD_SDK_VERSION) 