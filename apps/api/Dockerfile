FROM golang:1.26.5-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/url-shortener-api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/url-shortener-healthcheck ./cmd/healthcheck

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder /out/url-shortener-api /url-shortener-api
COPY --from=builder /out/url-shortener-healthcheck /url-shortener-healthcheck

EXPOSE 8080

USER nonroot:nonroot

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 CMD ["/url-shortener-healthcheck"]

ENTRYPOINT ["/url-shortener-api"]
