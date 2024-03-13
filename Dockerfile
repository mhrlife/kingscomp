FROM hub.hamdocker.ir/golang:1.21 as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/app .
ENTRYPOINT ["./app","serve"]