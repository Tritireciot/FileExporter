FROM golang:1.19-alpine AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download


COPY /internal/.. .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .


FROM ubuntu:22.04

WORKDIR /app


RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app .


RUN chmod +x ./app

CMD ["./app"]
