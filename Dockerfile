FROM golang:alpine

RUN apk add git

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN go build -o server /app/cmd/server/main.go

EXPOSE 8080

ENTRYPOINT [ "./server" ]
