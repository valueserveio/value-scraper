FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .


FROM scratch as production

# Install certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

COPY .env.local .
COPY --from=builder /app/main .

EXPOSE 3000

ENTRYPOINT ["./main"]