#!/bin/sh
set -e

export MINIO_ACCESS_KEY=FJDSJ
export MINIO_SECRET_KEY=DSG643HGDS

mkdir -p /tmp/minio
minio server /tmp/minio &>/dev/null &
sleep 5

go test ./pkg/mounter ./pkg/s3 ./cmd/s3driver -cover
go test ./pkg/driver -cover -ginkgo.noisySkippings=false -ginkgo.skip="should fail when requesting to create a volume with already existing name and different capacity"
