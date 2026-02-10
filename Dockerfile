FROM golang:1.19-alpine AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download


COPY /internal/.. .


RUN go build -o app .


FROM alpine:3.19

WORKDIR /app


RUN apk add --no-cache ca-certificates


COPY --from=builder /app/app .

CMD ["./app"]
