# --------Build stage--------
FROM golang:1.24-alpine AS builder

# install git
RUN apk add --no-cache git

WORKDIR /app

# cache module download
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY . .

# build static binary
RUN go build -o /myapp .

# --------Runtime Stage--------
FROM alpine:3.20

# non root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# copy compile binary from builder
COPY --from=builder /myapp /usr/local/bin/myapp

# Set working directory
WORKDIR /config

# non-root user
USER appuser

# Document the port listens on (e.g., 8080)
EXPOSE 11823

ENTRYPOINT [ "/usr/local/bin/myapp" ]
