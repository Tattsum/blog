FROM golang:1.24-bullseye AS builder

WORKDIR /app

# モジュールファイルのみコピーして依存関係をキャッシュ
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

# proto 生成用の Go プラグイン（gen/ は .gitignore のため CI で存在しない → ビルド時に生成）
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
ENV PATH="/go/bin:${PATH}"

# buf をインストール
ARG BUF_VERSION=1.32.0
RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" -o /usr/local/bin/buf \
    && chmod +x /usr/local/bin/buf

# ソースをコピー（gen/ はリポジトリに含まれないためこの時点では無い）
COPY . .

# proto から Go コードを生成（backend が gen/ を import するため必須。Go プラグインのみの設定を使用）
RUN buf generate --template buf.gen.go.yaml

# 静的バイナリをビルド（モノレポの backend/cmd/server）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./backend/cmd/server

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/server /app/server

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]

