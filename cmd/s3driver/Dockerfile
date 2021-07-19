FROM golang:1.16-alpine as gobuild

WORKDIR /build
ADD . /build

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

FROM debian:buster-slim
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"

# patched goofys
ADD goofys /usr/bin/

RUN apt-get update && \
  apt-get install -y ca-certificates && \
  rm -rf /var/lib/apt/lists/*

# s3fs curl unzip
## install rclone
#ARG RCLONE_VERSION=v1.54.1
#RUN cd /tmp \
#  && curl -O https://downloads.rclone.org/${RCLONE_VERSION}/rclone-${RCLONE_VERSION}-linux-amd64.zip \
#  && unzip /tmp/rclone-${RCLONE_VERSION}-linux-amd64.zip \
#  && mv /tmp/rclone-*-linux-amd64/rclone /usr/bin \
#  && rm -r /tmp/rclone*

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
