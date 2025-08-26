FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS gobuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
ADD go.mod go.sum /build/
RUN go mod download -x
ADD cmd /build/cmd
ADD pkg /build/pkg
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

FROM alpine:3.20
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"
ARG TARGETOS
ARG TARGETARCH
RUN apk add --no-cache fuse mailcap rclone s3fs-fuse
ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-${TARGETOS}-${TARGETARCH} /usr/bin/geesefs
RUN chmod 755 /usr/bin/geesefs

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
