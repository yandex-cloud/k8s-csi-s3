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

# Compile & install s3backer
RUN git clone https://github.com/archiecobbs/s3backer.git ./s3backer

WORKDIR "./s3backer"

RUN ./autogen.sh && \
    ./configure && \
    make && \
    make install

COPY ./_output/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
