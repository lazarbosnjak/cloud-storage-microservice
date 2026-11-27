# Stage 1: Build

FROM golang:1.26.10-alpine AS build

WORKDIR /backend

COPY ./sharing-service/go.mod ./sharing-service/go.sum ./sharing-service/

COPY ./proto ./proto

WORKDIR /backend/sharing-service

RUN go mod download

COPY ./sharing-service .

RUN go build -o sharing-service main.go

# Stage 2: Run
FROM alpine:3.24.4

WORKDIR /app

COPY --from=build /backend/sharing-service/sharing-service ./sharing-service

ENTRYPOINT [ "/app/sharing-service" ]
