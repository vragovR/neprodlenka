# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS dev
COPY certs/russian_trusted_root_ca.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["go", "run", "./cmd/neprodlenka", "run"]

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /neprodlenka ./cmd/neprodlenka

FROM alpine:3 AS prod
RUN apk add --no-cache tzdata ca-certificates
COPY certs/russian_trusted_root_ca.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
COPY --from=build /neprodlenka /usr/local/bin/neprodlenka
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["neprodlenka"]
CMD ["run"]
