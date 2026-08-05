# syntax=docker/dockerfile:1

FROM node:20-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY backend/ ./backend/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/house-backend ./backend

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app \
    && mkdir -p /app/dist /app/uploads /app/data \
    && chown -R app:app /app
WORKDIR /app
COPY --from=backend-builder /out/house-backend /app/house-backend
COPY --from=frontend-builder /src/frontend/dist /app/dist
USER app
ENV DATABASE_URL=/app/data/house.db
ENV CORS_ALLOWED_ORIGINS=*
EXPOSE 8080
CMD ["/app/house-backend"]
