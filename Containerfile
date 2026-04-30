# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
RUN apk update && apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p /app/public/uploads
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -v -o /bin/pvapi .

FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/pvapi /bin/pvapi
COPY --from=build /app/public/uploads /app/public/uploads

EXPOSE 8080
CMD ["/bin/pvapi"]
