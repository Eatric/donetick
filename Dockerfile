# syntax=docker/dockerfile:1.7

ARG NODE_VERSION=20-alpine
ARG GO_VERSION=1.24-alpine
ARG RUNTIME_VERSION=3.21

FROM node:${NODE_VERSION} AS frontend-builder

WORKDIR /src/frontend

COPY frontend-src/package.json frontend-src/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY frontend-src/ ./

ARG VITE_APP_API_URL=
ARG VITE_APP_REDIRECT_URL=https://task.iksanov.dev
ARG VITE_APP_GOOGLE_CLIENT_ID=USE_YOUR_OWN_CLIENT_ID
ARG VITE_OPENREPLAY_PROJECT_KEY=

ENV VITE_APP_API_URL=${VITE_APP_API_URL} \
    VITE_APP_REDIRECT_URL=${VITE_APP_REDIRECT_URL} \
    VITE_APP_GOOGLE_CLIENT_ID=${VITE_APP_GOOGLE_CLIENT_ID} \
    VITE_OPENREPLAY_PROJECT_KEY=${VITE_OPENREPLAY_PROJECT_KEY} \
    VITE_IS_SELF_HOSTED=true

RUN npm run build -- --mode selfhosted

FROM golang:${GO_VERSION} AS backend-builder

WORKDIR /src/backend

RUN apk add --no-cache build-base ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . ./
COPY --from=frontend-builder /src/frontend/dist ./frontend/dist

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -trimpath -ldflags='-s -w' -o /out/donetick .

FROM alpine:${RUNTIME_VERSION}

RUN apk add --no-cache ca-certificates libc6-compat sqlite tzdata

WORKDIR /

RUN install -d -m 0750 /donetick-data

COPY --from=backend-builder /out/donetick /donetick
COPY --from=backend-builder /src/backend/config /config

ENV DT_ENV=selfhosted \
    DT_SQLITE_PATH=/donetick-data/donetick.db

VOLUME ["/donetick-data"]

EXPOSE 2021

HEALTHCHECK --start-period=60s --interval=60s --timeout=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:2021/api/v1/health || exit 1

CMD ["/donetick"]
