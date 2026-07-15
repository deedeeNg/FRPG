# --- Build stage ---
FROM golang:1.22-alpine AS build
WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# --- Runtime stage ---
FROM alpine:3.20
WORKDIR /app

# AWS Lambda Web Adapter: lets the unmodified net/http server run on Lambda
# (see DEPLOYMENT.md). Lambda auto-loads binaries under /opt/extensions; OUTSIDE
# Lambda this file is simply never invoked, so the SAME image still runs under
# docker-compose / App Runner / locally with no behavior change. The readiness
# path points at a route the mux actually serves (default "/" would 404).
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 /lambda-adapter /opt/extensions/lambda-adapter
ENV AWS_LWA_READINESS_CHECK_PATH=/api/health

COPY --from=build /server /server
EXPOSE 8080
ENV PORT=8080
CMD ["/server"]
