FROM golang:1.25-bookworm

ENV GOTOOLCHAIN=auto
ENV PATH="/usr/local/go/bin:/go/bin:${PATH}"
ENV GOFLAGS="-buildvcs=false"

WORKDIR /workspace
COPY go.mod go.sum ./
RUN /usr/local/go/bin/go mod download

RUN /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@v0.38.0
