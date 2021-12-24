FROM golang:1.17-alpine3.14 AS build

WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download
COPY . /app
RUN go build -ldflags="-s -w" -o /app/aim
RUN chmod +x /app/aim

FROM golang:1.17-alpine3.14 AS prod

WORKDIR /app

EXPOSE 5190
ARG OSCAR_HOST
ARG OSCAR_PORT
ARG OSCAR_BOS_HOST
ARG OSCAR_BOS_PORT

COPY --from=build /app/models /app/models
COPY --from=build /app/aim /app/aim
CMD ["/app/aim"]
