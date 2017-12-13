FROM golang:1.9

RUN go get -u github.com/golang/dep/cmd/dep && go install github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/icedmocha/rss-client
COPY . /go/src/github.com/icedmocha/rss-client

RUN dep ensure && go install

ENTRYPOINT ["rss-client"]
