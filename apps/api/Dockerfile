FROM golang:1.23.5-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/url-shortener-api ./cmd/api

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder /out/url-shortener-api /url-shortener-api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/url-shortener-api"]
