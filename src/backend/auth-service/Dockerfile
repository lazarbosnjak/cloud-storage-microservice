# Stage 1: Build

FROM golang:1.26.10-alpine AS build

WORKDIR /backend

COPY ./auth-service/go.mod ./auth-service/go.sum ./auth-service/

COPY ./proto ./proto

WORKDIR /backend/auth-service

RUN go mod download

COPY ./auth-service .

RUN go build -o auth-service main.go

# Stage 2: Run
FROM alpine:3.24.4

WORKDIR /app

COPY --from=build /backend/auth-service/auth-service ./auth-service

ENTRYPOINT [ "/app/auth-service" ]
