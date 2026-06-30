# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/server ./cmd/server

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

# distroless static includes ca-certificates and a non-root user "nonroot".
USER nonroot:nonroot

WORKDIR /app
COPY --from=build /out/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
