FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o httpflood .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/httpflood ./httpflood
EXPOSE 8080
ENTRYPOINT ["./httpflood", "serve"]
