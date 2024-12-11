#FROM golang:1.19-alpine as gobuild
FROM registry.cn-hangzhou.aliyuncs.com/eryajf/golang:1.20.14-alpine3.19 as gobuild

WORKDIR /build
ENV GOPROXY=https://goproxy.cn
ADD go.mod go.sum /build/
RUN go mod download -x
ADD cmd /build/cmd
ADD pkg /build/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

#FROM alpine:3.17
FROM registry.cn-hangzhou.aliyuncs.com/eryajf/alpine:3.19
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"

RUN apk add --no-cache fuse mailcap rclone
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/community s3fs-fuse

#ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-amd64 /usr/bin/geesefs
ADD geesefs /usr/bin/geesefs
RUN chmod 755 /usr/bin/geesefs

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
