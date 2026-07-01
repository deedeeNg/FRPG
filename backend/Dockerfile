# --- Build stage ---
FROM golang:1.22-alpine AS build
WORKDIR /app

# Cache dependencies first
COPY go.mod ./
# COPY go.sum ./   # add back once external deps are introduced
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# --- Runtime stage ---
FROM alpine:3.20
WORKDIR /app
COPY --from=build /server /server
EXPOSE 8080
ENV PORT=8080
CMD ["/server"]
