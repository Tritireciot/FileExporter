FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY . .
COPY go.mod go.sum ./
RUN go mod download


RUN CGO_ENABLED=0 GOOS=linux go build -o app ./PrintServer


FROM ubuntu:22.04

WORKDIR /app


RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    fontconfig \
    libfontconfig1 \
    fonts-liberation \
    fonts-noto-color-emoji \
    && rm -rf /var/lib/apt/lists/*

ENV FONTCONFIG_PATH=/etc/fonts

COPY --from=builder /app/app .


RUN chmod +x ./app

CMD ["./app"]
