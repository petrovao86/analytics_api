FROM golang:1.23.4-bookworm AS build

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY . .
RUN go build -o /service ./cmd/

FROM debian:bookworm-slim

RUN adduser --system --home /app --uid 1000 --group app
USER 1000

WORKDIR /app/

COPY --from=build --chown=app:app /service /app/service
EXPOSE 8888
ENTRYPOINT [ "/app/service"]
