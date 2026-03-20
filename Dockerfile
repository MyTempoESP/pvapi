# syntax=docker/dockerfile:1
FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN mkdir -p /app/public/uploads
RUN CGO_ENABLED=0 go build -o /bin/pvapi ./pvapi

FROM scratch
WORKDIR /app
COPY --from=build /bin/pvapi /bin/pvapi
COPY --from=build /app/public/uploads /app/public/uploads

EXPOSE 8080
CMD ["/bin/pvapi"]
