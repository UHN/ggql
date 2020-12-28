FROM golang:1.15

WORKDIR /go/src/market
COPY main.go go.mod price.graphql ./

ENV GO111MODULE=on
RUN go build .

EXPOSE 3000
CMD ./market
