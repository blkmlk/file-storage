FROM golang:1.19-alpine as builder

WORKDIR /app
ADD . .

RUN go build -v -o /out cmd/storage/main.go

FROM alpine:3.15

COPY --from=builder /out /app

CMD ["/app"]
