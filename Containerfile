# syntax=docker/dockerfile:1
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p /app/public/uploads
RUN CGO_ENABLED=0 go build -v -o /bin/pvapi .

FROM scratch
WORKDIR /app
COPY --from=build /bin/pvapi /bin/pvapi
COPY --from=build /app/public/uploads /app/public/uploads

EXPOSE 8080
CMD ["/bin/pvapi"]
