FROM golang:1.19-alpine as gobuild

WORKDIR /build
ADD go.mod go.sum /build/
RUN go mod download -x
ADD cmd /build/cmd
ADD pkg /build/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

FROM alpine:3.17
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"

RUN apk add --no-cache fuse
#RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/community rclone s3fs-fuse

ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-amd64 /var/lib/kubelet/plugins/ru.yandex.s3.csi/geesefs
RUN chmod 755 /var/lib/kubelet/plugins/ru.yandex.s3.csi/geesefs

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
