FROM ctrox/csi-s3:1.0.0-alpha
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="csi-s3 testing image"

RUN apt-get update && \
    apt-get install -y \
      git wget && \
    rm -rf /var/lib/apt/lists/*

RUN wget -q https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz && \
    tar -xf go1.10.3.linux-amd64.tar.gz && \
    mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

RUN go get -u github.com/minio/minio && go install github.com/minio/minio/cmd

ENTRYPOINT ["/go/src/github.com/ctrox/csi-s3/test/test.sh"]
