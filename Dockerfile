FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/banner-server ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/banner-client ./cmd/client

FROM alpine:3.22 AS runtime
RUN addgroup -S app && adduser -S -G app -u 10001 app
WORKDIR /app
COPY --from=builder /out/banner-server /app/banner-server
COPY --from=builder /out/banner-client /app/banner-client
COPY --from=builder /src/rules /app/rules
COPY --from=builder /src/testdata /app/testdata
USER 10001:10001
ENTRYPOINT ["/app/banner-server"]
