FROM golang:1.27 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o main ./cmd

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/main /
CMD ["/main"]
