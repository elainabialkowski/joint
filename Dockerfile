FROM golang:alpine AS build

WORKDIR /build

COPY . .

RUN go mod download
RUN go mod tidy
RUN go test -race ./...

RUN go build -o /build/server .

FROM scratch AS runner

COPY --from=build /build/server /server

ENTRYPOINT [ "/server" ]