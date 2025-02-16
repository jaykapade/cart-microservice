FROM golang:1.23-alpine3.20 AS build
RUN apk --no-cache add gcc g++ make ca-certificates
WORKDIR /go/src/github.com/jaykapade/cart-microservice
COPY go.mod go.sum ./
# COPY vendor vendor
# COPY account account
# COPY catalog catalog
# COPY order order
COPY . .
RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./order/cmd/order

FROM alpine:3.20
WORKDIR /usr/bin
COPY --from=build /go/bin .
EXPOSE 8080
CMD ["app"]