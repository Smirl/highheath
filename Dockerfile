# build hugo
FROM klakegg/hugo:alpine-onbuild AS hugo

#build stage
FROM golang:alpine AS builder
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
RUN go build -o /go/bin/app highheath/quickstart.go

#final stage
FROM alpine:latest
COPY --from=hugo /target /public
COPY --from=builder /go/bin/app /app
ENTRYPOINT ./app
EXPOSE 8080
