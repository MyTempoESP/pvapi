# syntax=docker/dockerfile:1
FROM golang:1.25
WORKDIR /app
COPY . .
RUN go build -o /bin/pvapi ./pvapi

FROM scratch
COPY --from=0 /bin/pvapi /bin/pvapi
CMD ["/bin/pvapi"]
