# Stage 1: Build

FROM golang:1.26.10-alpine AS build

WORKDIR /backend

COPY ./notification-service/go.mod ./notification-service/go.sum ./notification-service/

COPY ./proto ./proto

WORKDIR /backend/notification-service

RUN go mod download

COPY ./notification-service .

RUN go build -o notification-service main.go

# Stage 2: Run
FROM alpine:3.24.4

WORKDIR /app

COPY --from=build /backend/notification-service/notification-service ./notification-service

ENTRYPOINT [ "/app/notification-service" ]
