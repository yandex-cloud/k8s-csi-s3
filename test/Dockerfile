FROM ctrox/csi-s3:dev-full
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="csi-s3 testing image"

RUN apt-get update && \
  apt-get install -y \
  git wget make && \
  rm -rf /var/lib/apt/lists/*

ARG GOVERSION=1.16.3
RUN wget -q https://golang.org/dl/go${GOVERSION}.linux-amd64.tar.gz && \
  tar -xf go${GOVERSION}.linux-amd64.tar.gz && \
  rm go${GOVERSION}.linux-amd64.tar.gz && \
  mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

RUN wget -q https://dl.min.io/server/minio/release/linux-amd64/minio && \
  chmod +x minio &&\
  mv minio /usr/local/bin

WORKDIR /app

# prewarm go mod cache
COPY go.mod .
COPY go.sum .
RUN go mod download

ADD test/test.sh /usr/local/bin

ENTRYPOINT ["/usr/local/bin/test.sh"]
