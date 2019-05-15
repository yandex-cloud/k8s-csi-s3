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

FROM debian:stretch
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="csi-s3 image"
COPY --from=s3backer /usr/bin/s3backer /usr/bin/s3backer

# s3fs and some other dependencies
RUN apt-get update && \
  apt-get install -y \
  libfuse2 gcc sqlite3 libsqlite3-dev \
  s3fs psmisc procps libcurl3 xfsprogs curl unzip && \
  rm -rf /var/lib/apt/lists/*

# install rclone
ARG RCLONE_VERSION=v1.47.0
RUN cd /tmp \
  && curl -O https://downloads.rclone.org/${RCLONE_VERSION}/rclone-${RCLONE_VERSION}-linux-amd64.zip \
  && unzip /tmp/rclone-${RCLONE_VERSION}-linux-amd64.zip \
  && mv /tmp/rclone-*-linux-amd64/rclone /usr/bin \
  && rm -r /tmp/rclone*

COPY ./_output/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
