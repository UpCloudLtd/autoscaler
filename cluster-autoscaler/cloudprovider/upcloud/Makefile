
UPCLOUD_SDK_PACKAGE ?= github.com/upcloudltd/upcloud-go-api/v6
UPCLOUD_SDK_VERSION ?= v6.5.0

IMAGE_TAG ?= upcloud-autoscaler:0.1-dev

build:
	BUILD_TAGS=upcloud make -C ../../ build-in-docker

image: build
	docker build -t $(IMAGE_TAG) -f ../../Dockerfile.amd64 ../../

vendor:
	sh vendor.sh $(UPCLOUD_SDK_PACKAGE) $(UPCLOUD_SDK_VERSION) 
