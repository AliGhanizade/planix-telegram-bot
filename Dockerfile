FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /planix ./cmd/planix

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /planix /planix
EXPOSE 8080
ENTRYPOINT ["/planix"]
