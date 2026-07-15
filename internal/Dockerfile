FROM golang:bookworm
WORKDIR /src
# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

 RUN apt-get update && apt-get install -y \
    vim \
    nano \
    less \
    curl \
    netcat-openbsd \
 && rm -rf /var/lib/apt/lists/*
# Then copy the rest
COPY . .
# Force Linux amd64 build
RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/distributed_kvs ./cmd/distributed-kvs && \
    chmod +x /go/bin/distributed_kvs

# Add grpc_health_probe
ARG GRPC_HEALTH_PROBE_VERSION=v0.3.2
RUN wget -qO /bin/grpc_health_probe \
    "https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64" \
    && chmod +x /bin/grpc_health_probe


# COPY  /go/bin/distributed_kvs /bin/distributed_kvs
# COPY  /go/bin/grpc_health_probe /bin/grpc_health_probe
CMD ["sleep", "infinity"]
