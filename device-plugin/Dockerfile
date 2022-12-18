# Package go application as docker image
FROM golang:1.19 AS builder
WORKDIR /go/src/github.com/keyval-dev/odigos/device-plugin
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/device-plugin ./cmd/main.go

FROM gcr.io/distroless/static
COPY --from=builder /go/bin/device-plugin /
ENTRYPOINT ["/device-plugin"]