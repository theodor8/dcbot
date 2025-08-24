# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o dcbot

FROM gcr.io/distroless/static
COPY --from=builder /app/dcbot /dcbot
USER nonroot:nonroot
ENTRYPOINT ["/dcbot"]
