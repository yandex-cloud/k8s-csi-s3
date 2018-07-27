FROM golang:stretch
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="s3 fuse csi plugin"

RUN apt-get update && apt-get install -y \
  s3fs \
  build-essential \
  autoconf \
  libcurl4-openssl-dev \
  libfuse-dev \
  libexpat1-dev \
  libssl-dev \
  zlib1g-dev \
  xfsprogs \
  psmisc \
  git && \
  rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/archiecobbs/s3backer.git ./s3backer

WORKDIR "./s3backer"

RUN ["./autogen.sh"]
RUN ["./configure"]
RUN ["make"]
RUN ["make", "install"]

RUN go get -u github.com/minio/minio && go install github.com/minio/minio/cmd

CMD ["/go/src/github.com/ctrox/csi-s3-driver/test/test.sh"]
