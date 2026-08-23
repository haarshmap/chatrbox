FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN  CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o chatrbox ./cmd/

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/chatrbox .
EXPOSE 8080

CMD [ "./chatrbox" ]