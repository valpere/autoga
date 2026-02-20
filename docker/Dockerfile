FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /autoga ./cmd/autoga


FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /autoga /autoga

EXPOSE 8080

ENTRYPOINT ["/autoga"]
