# A Few notes:
# - Make sure the version you use for a builder belongs to the same minor release as the image you use to run the game
# - Don't use static linking! No really, static linking is bad from a secuirty point of view.
FROM golang:1.13.4-alpine3.10 as builder
ADD . /go/src/tobw
WORKDIR /go/src/tobw
RUN go build -v .

FROM alpine:3.10.3
RUN adduser -S -D -H -h /app tobw
USER tobw
COPY --from=builder /go/src/tobw/tobw /app/
WORKDIR /app
CMD ["./tobw"]