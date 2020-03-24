FROM alpine:3.11

RUN set -eux; \
    apk --no-cache upgrade; \
    apk add --no-cache ca-certificates tzdata

# Use /data volume for MCL and server data
RUN mkdir /data && chmod a+rwx /data
VOLUME [ "/data" ]
WORKDIR /data

# Copy binary from builder
COPY mcl /usr/local/bin/mcl

ENTRYPOINT [ "/usr/local/bin/mcl" ]
