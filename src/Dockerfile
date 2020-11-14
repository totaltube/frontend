# syntax=docker/dockerfile:experimental
FROM golang:1.15.3-alpine3.12 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* .
RUN go mod download
COPY . .

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
GOOS=linux GOARCH=amd64 go build -o /totaltube-frontend .

FROM alpine3.12 AS bin-unix
COPY --from=build /totaltube-frontend /
CMD ["/totaltube-frontend", "-c", "/config.toml"]