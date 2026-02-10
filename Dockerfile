# ---------- BUILD STAGE ----------
FROM golang:1.19-alpine AS builder

WORKDIR /app

# зависимости (кешируется)
COPY go.mod go.sum ./
RUN go mod download

# код
COPY . .

# сборка
RUN go build -o app .

# ---------- RUNTIME STAGE ----------
FROM alpine:3.19

WORKDIR /app

# сертификаты (важно для https / pg)
RUN apk add --no-cache ca-certificates

#COPY templates templates
# бинарь
COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
