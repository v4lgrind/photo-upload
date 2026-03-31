FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod main.go ./
RUN go build -o photo-upload .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/photo-upload .
COPY static/ static/
EXPOSE 8080
CMD ["./photo-upload"]
