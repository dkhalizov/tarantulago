FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY /src .

RUN go build -o /bin/app ./cmd/bot/main.go

FROM gcr.io/distroless/static
COPY --from=builder /bin/app /app
ENTRYPOINT ["/app"]