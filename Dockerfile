FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
RUN mkdir -p /app/storage && chown -R app:app /app

COPY --from=build /out/api /app/api

USER app

EXPOSE 8080

CMD ["/app/api"]
