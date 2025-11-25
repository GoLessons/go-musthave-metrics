FROM golang:1.25-bookworm AS base

ENV GOTOOLCHAIN=auto
ENV PATH="/usr/local/go/bin:/go/bin:${PATH}"
ENV GOFLAGS="-buildvcs=false"

WORKDIR /workspace
RUN apt-get update && apt-get install -y protobuf-compiler
RUN /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@v0.38.0
RUN /usr/local/go/bin/go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN /usr/local/go/bin/go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN ln -sf /go/bin/protoc-gen-go /usr/local/bin/protoc-gen-go && ln -sf /go/bin/protoc-gen-go-grpc /usr/local/bin/protoc-gen-go-grpc
RUN ln -sf /go/bin/protoc-gen-go /usr/local/bin/protoc-gen-go && ln -sf /go/bin/protoc-gen-go-grpc /usr/local/bin/protoc-gen-go-grpc
RUN /usr/local/go/bin/go install golang.org/x/tools/cmd/gofmt@latest || true
RUN /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@v0.38.0 || true
COPY go.mod go.sum ./
RUN /usr/local/go/bin/go mod download

FROM base AS grpcserver-builder
COPY . .
RUN mkdir -p internal/proto && protoc -I ./var/proto --go_out=paths=source_relative:./internal/proto --go-grpc_out=paths=source_relative:./internal/proto ./var/proto/metrics.proto
RUN /usr/local/go/bin/go mod tidy
RUN /usr/local/go/bin/go build -buildvcs=false -o /bin/grpcserver ./cmd/grpcserver

FROM base AS server-builder
COPY . .
RUN /usr/local/go/bin/go build -buildvcs=false -o /bin/server ./cmd/server

FROM golang:1.25-bookworm AS grpcserver-runtime
COPY --from=grpcserver-builder /bin/grpcserver /bin/grpcserver
ENTRYPOINT ["/bin/grpcserver"]

FROM golang:1.25-bookworm AS server-runtime
COPY --from=server-builder /bin/server /bin/server
ENTRYPOINT ["/bin/server"]

FROM base AS dev
COPY . .
RUN mkdir -p internal/proto && protoc -I ./var/proto --go_out=paths=source_relative:./internal/proto --go-grpc_out=paths=source_relative:./internal/proto ./var/proto/metrics.proto
RUN /usr/local/go/bin/go mod tidy
ENTRYPOINT ["/bin/bash", "-lc"]