FROM python:3.6 as s3ql-deps
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="s3ql dependencies"

RUN apt-get update && \
    apt-get install -y \
       python3 python3-setuptools \
      python3-dev python3-pip pkg-config cython \
      libfuse-dev libattr1-dev && \
    rm -rf /var/lib/apt/lists/*

RUN pip3 install llfuse apsw defusedxml dugong requests pycrypto

FROM debian:stretch as s3backer
ARG S3BACKER_VERSION=1.5.0

RUN apt-get update && apt-get install -y \
  build-essential \
  autoconf \
  libcurl4-openssl-dev \
  libfuse-dev \
  libexpat1-dev \
  libssl-dev \
  zlib1g-dev \
  psmisc \
  pkg-config \
  git && \
  rm -rf /var/lib/apt/lists/*

# Compile & install s3backer
RUN git clone https://github.com/archiecobbs/s3backer.git /src/s3backer
WORKDIR /src/s3backer
RUN git checkout tags/${S3BACKER_VERSION}

RUN ./autogen.sh && \
    ./configure && \
    make && \
    make install

FROM python:3.6-slim
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="csi-s3 production image"

RUN apt-get update && \
    apt-get install -y \
      libfuse2 gcc sqlite3 libsqlite3-dev \
      s3fs psmisc procps libcurl3 xfsprogs && \
    rm -rf /var/lib/apt/lists/*

ARG S3QL_VERSION=2.29
ENV S3QL_URL=https://github.com/s3ql/s3ql/releases/download/release-${S3QL_VERSION}/s3ql-${S3QL_VERSION}.tar.bz2

COPY --from=s3ql-deps /root/.cache /root/.cache
COPY --from=s3ql-deps /usr/local/lib/python3.6/site-packages /usr/local/lib/python3.6/site-packages
RUN pip install ${S3QL_URL} && rm -rf /root/.cache

COPY --from=s3backer /usr/bin/s3backer /usr/bin/s3backer

COPY ./_output/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
