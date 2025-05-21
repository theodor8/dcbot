# syntax=docker/dockerfile:1

FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o bot

FROM gcr.io/distroless/static
COPY --from=builder /app/bot /bot
USER nonroot:nonroot
ENTRYPOINT ["/bot"]
