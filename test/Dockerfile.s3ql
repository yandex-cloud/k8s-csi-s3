FROM golang:stretch
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="s3 fuse csi plugin"
ARG S3QL_VERSION=release-2.28

RUN apt-get update && \
    apt-get install -y \
      s3fs wget psmisc procps python3 python3-setuptools \
      python3-dev python3-pip python3-llfuse pkg-config \
      sqlite3 libsqlite3-dev python3-apsw cython && \
    rm -rf /var/lib/apt/lists/*

RUN pip3 install defusedxml dugong requests pycrypto

WORKDIR /usr/src
RUN wget -q https://github.com/s3ql/s3ql/archive/${S3QL_VERSION}.tar.gz
RUN tar -xzf ${S3QL_VERSION}.tar.gz
WORKDIR /usr/src/s3ql-${S3QL_VERSION}
RUN python3 setup.py build_cython build_ext --inplace
RUN python3 setup.py install

RUN go get -u github.com/minio/minio && go install github.com/minio/minio/cmd

CMD ["/go/src/github.com/ctrox/csi-s3-driver/test/test.sh"]
