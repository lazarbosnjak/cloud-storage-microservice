# Stage 1: Build

FROM golang:1.26.10-alpine AS build

WORKDIR /backend

COPY ./file-service/go.mod ./file-service/go.sum ./file-service/

COPY ./proto ./proto

WORKDIR /backend/file-service

RUN go mod download

COPY ./file-service .

RUN go build -o file-service main.go

# Stage 2: Run
FROM alpine:3.24.4

WORKDIR /app

COPY --from=build /backend/file-service/file-service ./file-service

ENTRYPOINT [ "/app/file-service" ]
