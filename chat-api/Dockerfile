# syntax=docker/dockerfile:1

FROM golang:1.23.3 AS build
WORKDIR /src

# Сначала зависимости
COPY go.mod go.sum ./
RUN go mod download

# Потом исходники
COPY . .

# Билдим из ./cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/chat-api ./cmd

FROM alpine:3.20
RUN adduser -D -H -u 10001 app && apk add --no-cache ca-certificates tzdata
COPY --from=build /bin/chat-api /bin/chat-api
USER app

# GRPC порт
EXPOSE 50058
ENV GRPC_ADDR=:50058

ENTRYPOINT ["/bin/chat-api"]
