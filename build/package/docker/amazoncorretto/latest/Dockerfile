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
FROM amazoncorretto

# Create an unprivelaged user for running MCL and a data volume owned by that
# user. The Amazon Linux base image does not include common tools like groupadd
# useradd. Instead, the user and group entries have to be added manually. The 
# user and group IDs are the defaults if groupadd and useradd were available.
RUN echo "mcl:x:998:" >> /etc/group && \
    echo "mcl:x:999:998::/home/mcl:/sbin/nologin" >> /etc/passwd && \
    mkdir /home/mcl && \
    mkdir /data && \
    chown mcl:mcl /data
VOLUME [ "/data" ]
WORKDIR /data

# Copy binary from builder
COPY --from=builder /mcl/mcl /usr/local/bin

USER mcl
ENTRYPOINT [ "/usr/local/bin/mcl" ]
CMD [ "--help" ]
