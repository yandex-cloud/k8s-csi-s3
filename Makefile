# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
.PHONY: test build container push clean

REGISTRY_NAME=cr.yandex/crp9ftr22d26age3hulg
IMAGE_NAME=csi-s3
VERSION ?= 0.30.7
IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE_NAME):$(VERSION)
TEST_IMAGE_TAG=$(IMAGE_NAME):test

build:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o _output/s3driver ./cmd/s3driver
test:
	docker build -t $(TEST_IMAGE_TAG) -f test/Dockerfile .
	docker run --rm --privileged -v $(PWD):/build --device /dev/fuse $(TEST_IMAGE_TAG)
container:
	docker build -t $(IMAGE_TAG) -f cmd/s3driver/Dockerfile .
push: container
	docker tag $(IMAGE_TAG) $(REGISTRY_NAME)/$(IMAGE_NAME):latest
	docker push $(IMAGE_TAG)
	docker push $(REGISTRY_NAME)/$(IMAGE_NAME)
clean:
	go clean -r -x
	-rm -rf _output
