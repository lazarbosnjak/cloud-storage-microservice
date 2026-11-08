# Stage 1: Build

FROM golang:1.26.10-alpine AS build

WORKDIR /backend

COPY ./analytics-service/go.mod ./analytics-service/go.sum ./analytics-service/

COPY ./proto ./proto

WORKDIR /backend/analytics-service

RUN go mod download

COPY ./analytics-service .

RUN go build -o analytics-service main.go

# Stage 2: Run
FROM alpine:3.24.4

WORKDIR /app

COPY --from=build /backend/analytics-service/analytics-service ./analytics-service

ENTRYPOINT [ "/app/analytics-service" ]
