# Stage 1: Build

FROM golang:1.26.10-alpine AS build

WORKDIR /backend

COPY ./api-gateway/go.mod ./api-gateway/go.sum ./api-gateway/

COPY ./proto ./proto

WORKDIR /backend/api-gateway

RUN go mod download

COPY ./api-gateway .

RUN go build -o api-gateway main.go

# Stage 2: Run
FROM alpine:3.24.4

WORKDIR /app

COPY --from=build /backend/api-gateway/api-gateway ./api-gateway

ENTRYPOINT [ "/app/api-gateway" ]
