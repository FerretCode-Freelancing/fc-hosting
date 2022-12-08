FROM golang:1.19-alpine
RUN useradd -u 6891 nonroot
USER nonroot
WORKDIR /usr/route
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . . 
CMD ["go", "run", "main.go"]
