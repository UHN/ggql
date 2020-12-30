FROM golang:1.15

WORKDIR /go/src/market
COPY main.go websoc.go frame.go go.mod price.html ./

ENV GO111MODULE=on
RUN go build .

EXPOSE 3000
CMD ./market
