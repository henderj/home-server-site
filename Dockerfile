FROM golang:1.24-bookworm

RUN apt-get update && apt-get install -y sqlite3 libsqlite3-dev pkg-config && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o website .

EXPOSE 8080

CMD ["./website"]
