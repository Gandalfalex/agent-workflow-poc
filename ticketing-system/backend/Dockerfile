FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

COPY backend ./backend
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -o /app/ticketing-api ./cmd/server

FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend ./
ARG BASE_PATH=/ticketing
ENV BASE_PATH=${BASE_PATH}
ARG VITE_API_BASE=
ENV VITE_API_BASE=${VITE_API_BASE}
RUN npm run build

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /app/ticketing-api /app/ticketing-api
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
COPY backend/migrations /app/backend/migrations

USER app
EXPOSE 8080
ENTRYPOINT ["/app/ticketing-api"]
