# syntax=docker/dockerfile:1

## Build
FROM golang:1.20 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY main.go ./
COPY pkg ./pkg

RUN go build -o /connection-exporter

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /connection-exporter /connection-exporter
COPY connection-exporter.yaml /etc/connection-exporter.yaml

EXPOSE 8144

USER nonroot:nonroot

ENTRYPOINT ["/connection-exporter"]
CMD [ "-config", "/etc/connection-exporter.yaml" ]
