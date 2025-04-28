FROM golang:1.18-alpine AS builder

WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod ./
RUN go mod download

# ソースコードのコピー
COPY . .

# アプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o authzen-server

# 最終イメージの作成
FROM alpine:latest

WORKDIR /app

# ビルドしたバイナリをコピー
COPY --from=builder /app/authzen-server .

# ポートの公開
EXPOSE 8080

# アプリケーションの実行
ENTRYPOINT ["/app/authzen-server"]
