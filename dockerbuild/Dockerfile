# Build Stage
FROM golang:1.22.5-alpine AS builder

WORKDIR /app

ADD . .

RUN GOCACHE=OFF \
    && apk add git \
    && go mod tidy \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o app main.go

# Final Stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .
COPY dockerbuild/entrypoint.sh /app/entrypoint.sh

RUN apk add --update --no-cache \
    && apk add --no-cache ca-certificates tzdata curl bash \
    && mkdir -p logs \
    && addgroup -S app \
    && adduser -H -D -S app -G app \
    && chown -R app:app /app

USER app

ENTRYPOINT [ "./entrypoint.sh" ]