FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -o lifx-mqtt ./cmd/lifx-mqtt


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/lifx-mqtt .
RUN chmod +x lifx-mqtt

ENTRYPOINT ["./lifx-mqtt"]
