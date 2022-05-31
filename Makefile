# This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
# © 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
#
# This AWS Content is provided subject to the terms of the AWS Customer Agreement
# available at http://aws.amazon.com/agreement or other written agreement between
# Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.

export GOPROXY=direct

# NOTE this is typically the AWS Account number (set the environment variable before running `make` to override
REGISTRY_ID ?= 123456789012
IMAGE_NAME ?= aws-virtual-kubelet
REGION ?= us-west-2

ARCH ?= amd64
PLATFORM ?= linux/$(ARCH)

LINTER_BIN ?= golangci-lint

GO111MODULE := on
export GO111MODULE

VK_CMD_DIR = cmd/virtual-kubelet

VKVMA_API_DIR = api/vkvmagent/v0
VKVMA_PROTO_DIR = proto/vkvmagent/v0
VKVMA_MOCKS_DIR = mocks/generated/vkvmagent/v0
VKVMA_PROTO_FILES = $(addprefix $(VKVMA_PROTO_DIR)/, application_lifecycle.pb.go bootstrap.pb.go application_lifecycle_grpc.pb.go bootstrap_grpc.pb.go)
VKVMA_MOCKS_FILES = $(addprefix $(VKVMA_MOCKS_DIR)/, mock_application_lifecycle_grpc.pb.go mock_bootstrap_grpc.pb.go)

GRPC_API_DIR = api/grpc/health/v1
GRPC_PROTO_DIR = proto/grpc/health/v1
GRPC_MOCKS_DIR = mocks/generated/grpc/health/v1
GRPC_PROTO_FILES = $(addprefix $(GRPC_PROTO_DIR)/, health.pb.go health_grpc.pb.go)
GRPC_MOCKS_FILES = $(addprefix $(GRPC_MOCKS_DIR)/, mock_health_grpc.pb.go)

.PHONY: default
default: help

.PHONY: help
help:
	@echo "Welcome to the virtual-kubelet Makefile help\n"

	@echo "Below are some common targets and an explanation of their function"
	@echo "Run them with 'make <target>' (e.g. 'make build')\n"
	@echo "make clean			Cleans build output and generated files"
	@echo "make build			Builds project and generates any missing files"
	@echo "make proto			Regenerates protobuf and grpc files from .proto definitions"
	@echo "make mocks			Regenerates mocks"
	@echo "make docker			Build docker image for virtual-kubelet"
	@echo "make push			Deploy docker image to ECR (set envars to configure)"

# clean build and generated paths/files
.PHONY: clean
clean: files := bin/virtual-kubelet proto/* mocks/generated/*
clean:
	@rm -r $(files) &>/dev/null || exit 0

# TODO put all PHONY entries on one line with explanatory comment
.PHONY: proto
# generate types (pb.go), gRPC client stub and server interface (pb_grpc.go)
proto: $(VKVMA_PROTO_FILES) $(GRPC_PROTO_FILES)
	$(info VKVMA_PROTO_FILES = $(VKVMA_PROTO_FILES))
	$(info GRPC_PROTO_FILES = $(GRPC_PROTO_FILES))

$(VKVMA_PROTO_DIR)/%.pb.go $(VKVMA_PROTO_DIR)/%_grpc.pb.go: $(VKVMA_API_DIR)/%.proto
	# NOTE the compiler `protoc` must be installed separately (see https://grpc.io/docs/protoc-installation/)
	protoc --go_opt=paths=source_relative -I api --go_out=proto --go-grpc_out=proto \
 			--go-grpc_opt=paths=source_relative $<

$(GRPC_PROTO_DIR)/%.pb.go $(GRPC_PROTO_DIR)/%_grpc.pb.go: $(GRPC_API_DIR)/%.proto
	protoc --go_opt=paths=source_relative -I api --go_out=proto --go-grpc_out=proto \
 			--go-grpc_opt=paths=source_relative $<

# generate mocks
.PHONY: mocks
mocks: $(VKVMA_MOCKS_FILES) $(GRPC_MOCKS_FILES)

$(VKVMA_MOCKS_DIR)/mock_%_grpc.pb.go: $(VKVMA_PROTO_DIR)/%_grpc.pb.go
	mockgen -source $< -destination $@

$(GRPC_MOCKS_DIR)/mock_%_grpc.pb.go: $(GRPC_PROTO_DIR)/%_grpc.pb.go
	mockgen -source $< -destination $@

.PHONY: build

build: proto bin/virtual-kubelet

.PHONY: test
test: $(VKVMA_MOCKS_FILES) $(GRPC_MOCKS_FILES)
	@echo running tests
	@go test -v -race ./...


IMAGE?=$(REGISTRY_ID).dkr.ecr.$(REGION).amazonaws.com/$(IMAGE_NAME)
TAG?=$(shell git describe --tags --always --dirty="-dev" | sed 's/+/-/g')

.PHONY: docker
docker: 
	@echo 'Building image $(IMAGE):$(TAG) ...'
	# Specify versions in Dockerfile instead of using --no-cache (a change to version text in Dockerfile will force
	#  a rebuild of that layer
	docker build --platform $(PLATFORM) --memory 4G -t $(IMAGE):$(TAG)-$(ARCH) -f ./build/ci/Dockerfile .

.PHONY: push
push: docker

	aws ecr get-login-password \
    --region $(REGION) \
	| docker login \
    --username AWS \
    --password-stdin $(REGISTRY_ID).dkr.ecr.$(REGION).amazonaws.com
	docker push $(IMAGE):$(TAG)-$(ARCH)

.PHONY: vet
vet:
	@go vet ./... #$(packages)

.PHONY: lint
lint:
	@$(LINTER_BIN) run --new-from-rev "HEAD~$(git rev-list main.. --count)" ./...

.PHONY: check-mod
check-mod: # verifies that module changes for go.mod and go.sum are checked in
	@hack/ci/check_mods.sh

.PHONY: mod
mod:
	@go mod tidy

bin/virtual-kubelet: BUILD_VERSION          ?= $(shell git describe --tags --always --dirty="-dev")
bin/virtual-kubelet: BUILD_DATE             ?= $(shell date -u '+%Y-%m-%d-%H:%M UTC')
bin/virtual-kubelet: VERSION_FLAGS    := -ldflags='-X "main.buildVersion=$(BUILD_VERSION)" -X "main.buildTime=$(BUILD_DATE)"'

# Produce a statically-linked binary (all dependencies in a single executable), which allows use of the "scratch" docker
#  image for a minimal yet portable image
bin/%:
	CGO_ENABLED=0 go build -mod vendor -ldflags '-extldflags "-static"' -o bin/$(*) $(VERSION_FLAGS) $(VK_CMD_DIR)/main.go

.PHONY: deploy-vk-instance-profile
deploy-vk-instance-profile:
	aws cloudformation deploy --template-file cloudformation/instance-role-and-policy.yml --stack-name VKInstanceRoleAndPolicy --capabilities CAPABILITY_NAMED_IAM

.PHONY: delete-vk-instance-profile
delete-vk-instance-profile:
	aws cloudformation delete-stack --stack-name VKInstanceRoleAndPolicy

