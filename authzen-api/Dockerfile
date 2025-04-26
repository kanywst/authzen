FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# アプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o authzen-api .

# 最終イメージの作成
FROM alpine:3.19

WORKDIR /app

# ビルドしたバイナリをコピー
COPY --from=builder /app/authzen-api .

# ポートの公開
EXPOSE 8080

# アプリケーションの実行
CMD ["./authzen-api"]
