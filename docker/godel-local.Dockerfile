# Define the builder stage
# Define the builder stage
# build stage
FROM debian:bookworm as builder

# install on-demand tools
RUN apt-get update && apt-get install -y \
    curl \
    git \
    wget \
    build-essential \
    && apt-get clean

# install Go 1.21
# Set Go version
ENV GO_VERSION=1.21.11

# Download and install Go
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# set go env.
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY hack/ hack/
COPY vendor/ vendor/
COPY Makefile Makefile
COPY Makefile.expansion Makefile.expansion

RUN export GO_BUILD_PLATFORMS=linux/amd64 && make build

FROM debian:bookworm
RUN apt-get update && \
    apt-get install -y binutils && \
    apt-get clean && \
    ldd --version

WORKDIR /root
COPY --from=builder /workspace/bin/linux_amd64/* /usr/local/bin/
