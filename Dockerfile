FROM golang:1.18 AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN GOOS=linux go build -o ./birgedo ./cmd/birgeDo/

FROM alpine:latest

WORKDIR /root/

COPY --from=0 /app/birgedo .

EXPOSE 4000

CMD ["ls", "./birgedo"]