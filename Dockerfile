# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.19-buster AS build

WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /warrant -v ./cmd/warrant/main.go

##
## Deploy
##
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /warrant /app/warrant-exec

EXPOSE 8000

USER warrant

WORKDIR /app

ENTRYPOINT ["./warrant-exec"]
