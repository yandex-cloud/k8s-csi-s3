FROM golang:1.16-alpine as gobuild

WORKDIR /build
ADD . /build

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

FROM debian:buster-slim
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="csi-s3 slim image"

# s3fs and some other dependencies
RUN apt-get update && \
  apt-get install -y \
  s3fs curl unzip && \
  rm -rf /var/lib/apt/lists/*

# install rclone
ARG RCLONE_VERSION=v1.47.0
RUN cd /tmp \
  && curl -O https://downloads.rclone.org/${RCLONE_VERSION}/rclone-${RCLONE_VERSION}-linux-amd64.zip \
  && unzip /tmp/rclone-${RCLONE_VERSION}-linux-amd64.zip \
  && mv /tmp/rclone-*-linux-amd64/rclone /usr/bin \
  && rm -r /tmp/rclone*

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
