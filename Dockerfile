FROM tianon/golang
ENV GOPATH /go
ADD . /go/src/github.com/orchardup/go-orchard
WORKDIR /go/src/github.com/orchardup/go-orchard
