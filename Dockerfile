# FROM golang:1.21.3-alpine3.18 AS builder
# RUN apk add --no-cache git

# # Build
# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod tidy
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux \
#     go build -a -installsuffix cgo -o myhealth-service cmd/app/main.go

# # Run
# FROM scratch
# WORKDIR /app
# COPY --from=builder /app/config /config
# COPY --from=builder /app/myhealth-service .
# EXPOSE 7081
# RUN ["chmod", "a+x", "/app/myhealth-service"]
# CMD ["/app/myhealth-service"]

# Step 1: Modules caching
FROM golang:1.21.3-alpine3.18 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.21.3-alpine3.18 as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/app ./cmd/app

# Step 3: Final
FROM scratch
COPY --from=builder /app/config /config
COPY --from=builder /bin/app /app
CMD ["/app"]