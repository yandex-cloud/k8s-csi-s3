ARG ALPINE_VERSION="3.17"
ARG GOLANG_VERSION="1.20"
FROM docker.io/library/golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as gobuild

WORKDIR /build
ADD go.mod go.sum /build/
ADD cmd /build/cmd
ADD pkg /build/pkg

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

FROM docker.io/library/alpine:${ALPINE_VERSION}
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"

# apk add temporarily broken:
#ERROR: unable to select packages:
#  so:libcrypto.so.3 (no such package):
#    required by: s3fs-fuse-1.91-r1[so:libcrypto.so.3]
#RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing s3fs-fuse rclone

ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-amd64 /usr/bin/geesefs
RUN chmod 755 /usr/bin/geesefs

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
