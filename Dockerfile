# A Few notes:
# - Make sure the version you use for a builder belongs to the same minor release as the image you use to run the game
FROM golang:1.13.4-alpine3.10 as builder
ADD . /go/src/tobw
WORKDIR /go/src/tobw
RUN apk add --update make git && \
    make bootstrap && \
    make dep && \
    make build

FROM alpine:3.10.3
RUN adduser -S -D -H -h /app tobw
USER tobw
COPY --from=builder /go/src/tobw/bin/tobw_linux_amd64 /app/tobw
WORKDIR /app
CMD ["./tobw"]
