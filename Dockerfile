# A dockerfile for a memhashd image.
FROM golang:1.8-alpine
MAINTAINER Yasha Bubnov <girokompass@gmail.com>

# Copy the source code of the project (quite ugly but allows to build
# and compile everything in a single pass).
COPY . /go/src/memhashd

# Compile the project and remove the source code from the previous layer.
RUN go build -o /usr/bin/memhashd memhashd \
    && rm -rf /go/src/memhashd

RUN mkdir -p /etc/memhash.d/ /var/lib/memhashd/
COPY certs /etc/memhash.d/
WORKDIR /var/lib/memhashd

ENTRYPOINT ["/usr/bin/memhashd"]
CMD [""]
