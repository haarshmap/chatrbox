#For the build image
FROM golang:1.27 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/go.mod go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o main ./cmd

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/main /

#For the runtime image

ENV env=production
WORKDIR /app

USER 65532:65532
COPY --link --from=builder --chown=65532:65532 /app/main /main

EXPOSE 8080

CMD ["/main"]
