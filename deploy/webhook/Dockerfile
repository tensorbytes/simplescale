FROM golang:1.17.6 as builder

COPY . /app
WORKDIR /app

ENV GOPROXY=https://goproxy.io,direct
RUN GOOS=linux CGO_ENABLED=0 go build -o scale-webhook ./cmd/webhook/main.go

FROM alpine:3.15

COPY --from=builder /app/scale-webhook /app/scale-webhook

CMD ["/app/scale-webhook"]