# The build stage
FROM golang:1.22.2 as builder

WORKDIR /app
COPY . . 

# Pobierz narzędzie `migrate`
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xz && \
    mv migrate /usr/local/bin/

# Skonfiguruj `make`
RUN apt-get update && apt-get install -y make

# Budowanie aplikacji
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/*.go

# The run stage
FROM alpine:latest
WORKDIR /app

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/api .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/
COPY --from=builder /app/Makefile /app/

# Pobierz make (jeśli Alpine)
RUN apk add --no-cache make

# Uruchom migracje przed startem aplikacji
RUN make migrate-up || echo "Migration failed but continuing..."

EXPOSE 8080
CMD ["./api"]