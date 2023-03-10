FROM golang

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o birgedo ./cmd/birgeDo/

CMD ["./birgedo"]