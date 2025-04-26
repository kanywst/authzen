package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kanywst/authzen-api/api"
	"github.com/kanywst/authzen-api/policy"
	"google.golang.org/grpc"
)

func main() {
	// コマンドライン引数の解析
	var (
		httpPort = flag.Int("http-port", 8080, "HTTP server port")
		grpcPort = flag.Int("grpc-port", 9000, "gRPC server port")
		mode     = flag.String("mode", "both", "Server mode: http, grpc, or both")
	)
	flag.Parse()

	// ポリシーストアの初期化
	store := policy.NewStore()

	// サンプルポリシーの追加
	store.AddPolicy("user:alice", "document:report", "read", true)
	store.AddPolicy("user:alice", "document:report", "write", true)
	store.AddPolicy("user:bob", "document:report", "read", true)
	store.AddPolicy("user:bob", "document:report", "write", false)
	store.AddPolicy("user:charlie", "document:report", "read", false)

	// Istio/Envoy統合用のサンプルポリシー
	store.AddPolicy("user:alice", "resource:/sample", "GET", true)
	store.AddPolicy("user:bob", "resource:/sample", "GET", true)
	store.AddPolicy("user:alice", "resource:/sample/admin", "GET", true)
	store.AddPolicy("user:bob", "resource:/sample/admin", "GET", false)

	// シグナル処理の設定
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// モードに応じてサーバーを起動
	switch *mode {
	case "http":
		startHTTPServer(store, *httpPort)
	case "grpc":
		startGRPCServer(store, *grpcPort)
	case "both":
		go startHTTPServer(store, *httpPort)
		go startGRPCServer(store, *grpcPort)
	default:
		log.Fatalf("不明なモード: %s", *mode)
	}

	// シグナルを待機
	sig := <-sigCh
	log.Printf("シグナルを受信しました: %v", sig)
	log.Println("サーバーをシャットダウンしています...")
}

// HTTP APIサーバーを起動
func startHTTPServer(store *policy.Store, port int) {
	// APIサーバーの初期化
	server := api.NewServer(store)

	// HTTPSサーバーの設定
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.Router(),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// 証明書ファイルの存在確認
	_, certErr := os.Stat("server.crt")
	_, keyErr := os.Stat("server.key")

	// サーバーの起動（証明書がある場合はHTTPS、ない場合はHTTP）
	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		log.Printf("証明書ファイルが見つかりません。HTTP Authorization APIサーバーを起動しています。ポート: %d", port)
		log.Printf("注意: 本番環境ではHTTPSを使用してください。")
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPサーバーの起動に失敗しました: %v", err)
		}
	} else {
		log.Printf("HTTPS Authorization APIサーバーを起動しています。ポート: %d", port)
		err := httpServer.ListenAndServeTLS("server.crt", "server.key")
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPSサーバーの起動に失敗しました: %v", err)
		}
	}
}

// gRPC APIサーバーを起動（Envoy External Authorization用）
func startGRPCServer(store *policy.Store, port int) {
	// gRPCサーバーの初期化
	grpcServer := grpc.NewServer()

	// Envoy認可サーバーの登録
	envoyAuthServer := api.NewEnvoyAuthServer(store)
	envoyAuthServer.Register(grpcServer)

	// リスナーの作成
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("gRPCリスナーの作成に失敗しました: %v", err)
	}

	// サーバーの起動
	log.Printf("gRPC Authorization APIサーバーを起動しています。ポート: %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPCサーバーの起動に失敗しました: %v", err)
	}
}
