FROM golang:1.24 AS builder
WORKDIR /src
COPY ../ .
RUN if [ ! -f go.mod ]; then \
      go mod init ethexporter && \
      go get github.com/ethereum/go-ethereum@latest; \
    fi
RUN go mod tidy
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags "-s -w" -o /out/ethexporter .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /out/ethexporter /ethexporter
USER nonroot:nonroot

EXPOSE 9100
ENTRYPOINT ["/ethexporter"]
