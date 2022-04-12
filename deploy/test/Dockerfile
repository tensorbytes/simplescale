FROM golang:1.17.6 as builder

COPY . /app
WORKDIR /app

ENV GOPROXY=https://goproxy.io,direct
RUN GOOS=linux CGO_ENABLED=0 go build -o random-metrics ./cmd/test-deployment/main.go

FROM alpine:latest

COPY --from=builder /app/random-metrics /app/random-metrics

CMD ["/app/random-metrics"]