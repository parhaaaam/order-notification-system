FROM golang:1.19 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-order-notification-system

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /docker-order-notification-system /docker-order-notification-system

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/docker-order-notification-system"]