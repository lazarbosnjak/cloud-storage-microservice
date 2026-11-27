# Stage 1: Build

FROM golang:1.26.10-alpine AS build

WORKDIR /backend

COPY ./user-service/go.mod ./user-service/go.sum ./user-service/

COPY ./proto ./proto

WORKDIR /backend/user-service

RUN go mod download

COPY ./user-service .

RUN go build -o user-service main.go

# Stage 2: Run
FROM alpine:3.24.4

WORKDIR /app

COPY --from=build /backend/user-service/user-service ./user-service

ENTRYPOINT [ "/app/user-service" ]
