FROM harbor.iportal.ru/hub.docker.com/library/golang:1.19.0-bullseye AS build

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY . .
RUN go build -o /worker ./cmd/worker

FROM harbor.iportal.ru/hub.docker.com/library/debian:bullseye-slim

RUN adduser --system --home /app --uid 1000 --group app
USER 1000

WORKDIR /app/

COPY --from=build --chown=app:app /worker /app/worker
EXPOSE 9100
CMD [ "/app/worker" ]
