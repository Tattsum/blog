FROM golang:1.23-bullseye AS builder

WORKDIR /app

# モジュールファイルのみコピーして依存関係をキャッシュ
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

# 残りのソースコードをコピー（モノレポルートから backend を含む）
COPY . .

# 静的バイナリをビルド（モノレポの backend/cmd/server）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./backend/cmd/server

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/server /app/server

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]

