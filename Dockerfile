FROM alpine:3.10.3
RUN mkdir -p /app/bin && mkdir -p /app/etc && mkdir -p /app/security && adduser -S -D -H -h /app tobw
COPY bin/tobw_linux_amd64 /app/bin/tobw
USER tobw
WORKDIR /app/bin
CMD ["/app/bin/tobw", "/app/etc/tobw.yaml"]
