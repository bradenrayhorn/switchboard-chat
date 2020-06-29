FROM golang:1.14.4 as build

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build .

FROM alpine:latest
COPY --from=build /app/switchboard-chat /app/

EXPOSE 8080

ENTRYPOINT ["/app/switchboard-chat"]
