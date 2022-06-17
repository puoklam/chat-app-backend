FROM golang:alpine as builder
ENV GO111MODULE=on
RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/main .

FROM scratch

# Copy the Pre-built binary file
COPY --from=builder /app/bin/main .
COPY --from=builder /app/ed25519 .

# Run executable
CMD ["./main"]
