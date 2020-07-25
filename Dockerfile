# build hugo
FROM klakegg/hugo:0.74.3-alpine AS hugo
COPY . /src
RUN hugo --minify --cleanDestinationDir

#build stage
FROM golang:alpine AS builder
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
RUN go build -o /go/bin/app cmd/highheath/main.go

#final stage
FROM alpine:latest
COPY --from=hugo /src/public /public
COPY --from=builder /go/bin/app /app
ENTRYPOINT ./app
EXPOSE 8080
