FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod main.go ./
RUN go build -o photo-upload .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/photo-upload .
COPY static/ static/
RUN addgroup -g 1000 app && adduser -S -u 1000 -G app app
USER app
EXPOSE 8080
CMD ["./photo-upload"]
