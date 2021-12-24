FROM golang:1.17-alpine3.14 AS build

WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download && go mod vendor
RUN ls -l /app

COPY . /app
RUN ls -l /app
RUN go build -ldflags="-s -w" -o aim
RUN chmod +x aim

FROM scratch AS prod

EXPOSE 5190
ARG OSCAR_HOST
ARG OSCAR_PORT
ARG OSCAR_BOS_HOST
ARG OSCAR_BOS_PORT

COPY --from=build /app/aim /app/aim
ENTRYPOINT /app/aim
