FROM golang:1.21.0-alpine
RUN adduser -D nonroot
USER nonroot
WORKDIR /usr/logger
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
CMD ["go", "run", "main.go"]