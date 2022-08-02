FROM golang:1.17.5 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o api


# final stage
FROM alpine:3.11.5
COPY --from=builder /app/api /app/