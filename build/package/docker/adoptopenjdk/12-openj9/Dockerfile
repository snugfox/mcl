#
# Use golang as builder image
#
FROM golang:1-alpine as builder

WORKDIR /mcl

# Download project and build dependencies
RUN apk add --no-cache ca-certificates git
COPY ./go.mod ./go.sum ./
RUN go mod download

# Copy and build MCL
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o mcl github.com/snugfox/mcl/cmd/mcl


#
# Create final image from builder
#
FROM adoptopenjdk/openjdk12-openj9:alpine-jre

# Create an unprivelaged user for running MCL and a data volume owned by that
# user. Adding the user and group first ensures that the UID and GID remain
# consistent regardless of other layers.
RUN addgroup -S mcl && \
    adduser -S -G mcl mcl && \
    mkdir /data && \
    chown mcl:mcl /data
VOLUME [ "/data" ]
WORKDIR /data

# Install application dependencies
RUN apk add --no-cache tzdata ca-certificates

# Copy binary from builder
COPY --from=builder /mcl/mcl /usr/local/bin

USER mcl
ENTRYPOINT [ "/usr/local/bin/mcl" ]
CMD [ "--help" ]
