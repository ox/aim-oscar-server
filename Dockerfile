FROM golang:1.17-alpine3.14 AS build

WORKDIR /app
COPY go.mod go.sum /app
RUN (([ ! -d "/app/vendor" ] && go mod download && go mod vendor) || true)

COPY . /app
RUN go build -ldflags="-s -w" -mod vendor -o "aim" main.go
RUN chmod +x "aim"

FROM scratch AS prod

EXPOSE 5190
ARG OSCAR_HOST
ARG OSCAR_PORT
ARG OSCAR_BOS_HOST
ARG OSCAR_BOS_PORT

COPY --from=build /app/aim /app/aim
ENTRYPOINT /app/aim
