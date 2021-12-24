from golang:1.17-alpine3.14

workdir /app
copy go.mod go.sum /app
run go mod download

copy . /app
run go build -o /app/aim

EXPOSE 5190
ARG OSCAR_HOST
ARG OSCAR_PORT
ARG OSCAR_BOS_HOST
ARG OSCAR_BOS_PORT

cmd /app/aim
