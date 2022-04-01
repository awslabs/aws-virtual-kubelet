# This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
# © 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
#
# This AWS Content is provided subject to the terms of the AWS Customer Agreement
# available at http://aws.amazon.com/agreement or other written agreement between
# Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.

FROM golang:1.16-alpine AS build

RUN addgroup -S vk && adduser -S vk -G vk

ENV GO111MODULE=on
ENV GOPROXY=direct

ENV GOOS=linux
ENV GOARCH=amd64

# M1 / Apple Silicon
#ENV GOOS=darwin
#ENV GOARCH=arm64

# Install tools required for project
# Run `docker build --no-cache .` to update dependencies
RUN apk add --update --no-cache make git protoc

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
WORKDIR /go/src/github.com/aws/aws-virtual-kubelet

COPY . ./

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

# Runing `make clean` to remove any pre-built binaries copied during COPY step and ensure a rebuild in docker environ
RUN make clean
RUN make proto
RUN make build

# This results in a single layer image
FROM scratch
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /go/src/github.com/aws/aws-virtual-kubelet/bin /bin

USER vk

CMD ["/bin/virtual-kubelet", "--provider", "ec2"]
