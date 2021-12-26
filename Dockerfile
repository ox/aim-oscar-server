FROM golang:1.17-alpine3.14 AS build

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
COPY go.mod go.sum /app
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/aim
RUN chmod +x /app/aim && chmod +rw /app/aim.db

FROM scratch AS prod

WORKDIR /app

EXPOSE 5190
ARG OSCAR_HOST
ARG OSCAR_PORT
ARG OSCAR_BOS_HOST
ARG OSCAR_BOS_PORT

# Import from builder.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /app/models /app/models
COPY --from=build /app/aim /app/aim
COPY --from=build /app/aim.db /app/aim.db

# Use an unprivileged user.
USER appuser:appuser

ENTRYPOINT ["/app/aim"]
