# Build stage
FROM golang:1.21 as builder

WORKDIR /workspace

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/

# Build the binder binary
RUN GOOS=linux GOARCH=amd64 go build -o bin/binder cmd/binder/main.go

# Build the scheduler binary
RUN GOOS=linux GOARCH=amd64 go build -o bin/scheduler cmd/scheduler/main.go

# Build the dispatcher binary
RUN GOOS=linux GOARCH=amd64 go build -o bin/dispatcher cmd/dispatcher/main.go

# Final stage
FROM debian:bookworm
RUN apt-get update && \
    apt-get install -y binutils && \
    apt-get clean && \
    ldd --version

WORKDIR /root
# Copy binaries from the builder stage
COPY --from=builder /workspace/bin/binder /usr/local/bin/
COPY --from=builder /workspace/bin/scheduler /usr/local/bin/
COPY --from=builder /workspace/bin/dispatcher /usr/local/bin/