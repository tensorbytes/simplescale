FROM golang:1.17.6 as builder

COPY . /app
WORKDIR /app

ENV GOPROXY=https://goproxy.io,direct
RUN GOOS=linux CGO_ENABLED=0 go build -o autoscaler ./cmd/simple-autoscaler/main.go

FROM alpine:latest

COPY --from=builder /app/autoscaler /app/autoscaler

CMD ["/app/autoscaler"]