FROM golang:1.4-cross

RUN mkdir -p /go/src/nginx-and
WORKDIR /go/src/nginx-and

ENV CROSSPLATFORMS \
        linux/amd64 linux/386 linux/arm \
        darwin/amd64 darwin/386 \
        freebsd/amd64 freebsd/386 freebsd/arm \
        windows/amd64 windows/386

ENV GOARM 5

COPY . /go/src/nginx-and

CMD set -x \
    && for platform in $CROSSPLATFORMS; do \
            GOOS=${platform%/*} \
            GOARCH=${platform##*/} \
                go build -v -o bin/nginx-and-${platform%/*}-${platform##*/}; \
    done
