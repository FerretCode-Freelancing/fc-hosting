FROM golang:1.19-alpine
RUN adduser -D nonroot \
  && echo "nonroot ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/nonroot \
  && chmod 0440 /etc/sudoers/nonroot
USER nonroot
WORKDIR /usr/logger
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
CMD ["go", "run", "main.go"]