FROM ctrox/csi-s3:1.0.1-alpha
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="csi-s3 testing image"

RUN apt-get update && \
    apt-get install -y \
      git wget make && \
    rm -rf /var/lib/apt/lists/*

RUN wget -q https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz && \
    tar -xf go1.10.3.linux-amd64.tar.gz && \
    rm go1.10.3.linux-amd64.tar.gz && \
    mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# install dep
RUN wget -q https://raw.githubusercontent.com/golang/dep/master/install.sh && \
    mkdir -p /go/bin && \
    sh ./install.sh

RUN go get -u github.com/minio/minio && go install github.com/minio/minio/cmd

ADD test/test.sh /usr/local/bin

ENTRYPOINT ["/usr/local/bin/test.sh"]
