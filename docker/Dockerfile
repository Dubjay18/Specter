FROM golang:1.22 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/specter ./cmd/specter
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/testserver ./cmd/testserver

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /out/specter /app/specter
COPY --from=builder /out/testserver /app/testserver
COPY internal/config/specter-docker.yaml /app/config/specter-docker.yaml

RUN mkdir -p /data && chown -R app:app /app /data
USER app

EXPOSE 8080 3000 3001 7946
