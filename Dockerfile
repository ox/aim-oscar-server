FROM golang:1.20.5-alpine3.18 AS build

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

ENV CGO_ENABLED=0
COPY . /app
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/aim-oscar-server
RUN chmod +x /app/aim-oscar-server

FROM golang:1.20.5-alpine3.18 AS prod

WORKDIR /app

EXPOSE 5190
EXPOSE 5191

# Import from builder.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /app/models /app/models
COPY --from=build /app/aim-oscar-server /app/aim-oscar-server

# Use an unprivileged user.
USER appuser:appuser

ARG config
COPY $config /etc/aim-oscar-server/config.yml

CMD ["/app/aim-oscar-server", "-config", "/etc/aim-oscar-server/config.yml"]

FROM prod as dev

ARG config
COPY $config /etc/aim-oscar-server/config.yml
CMD ["/app/aim-oscar-server", "-config", "/etc/aim-oscar-server/config.yml"]

FROM build as db_tools

WORKDIR /app

ARG config
COPY $config /etc/aim-oscar-server/config.yml

COPY . /app
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=amd64 go build -o migrate cmd/migrate/main.go

ENTRYPOINT ["/app/migrate", "-config", "/etc/aim-oscar-server/config.yml"]
