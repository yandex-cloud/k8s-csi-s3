# Golang can cross-compile
FROM --platform=$BUILDPLATFORM golang:1.19-alpine as gobuild

WORKDIR /build
ADD go.mod go.sum /build/
RUN go mod download -x
ADD cmd /build/cmd
ADD pkg /build/pkg

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

FROM --platform=$BUILDPLATFORM alpine:3.17 AS downloader

WORKDIR /work

RUN apk add --no-cache curl

ARG TARGETOS
ARG TARGETARCH
RUN curl https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-$TARGETOS-$TARGETARCH -LfSso /work/geesefs

FROM alpine:3.17
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"

RUN apk add --no-cache fuse mailcap rclone
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/community s3fs-fuse

COPY --from=downloader /work/geesefs /usr/bin/geesefs
RUN chmod 755 /usr/bin/geesefs

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
