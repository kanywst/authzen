# AuthZEN サンプルアプリケーション

このリポジトリには、AuthZEN（Authorization API）仕様に準拠したサンプル実装が含まれています。このサンプルアプリケーションは、Go言語で実装されており、Kubernetes（Kind）上で動作確認できるように設計されています。

## 概要

```mermaid
graph TD
    A[クライアント] --> B[AuthZEN API サーバー]
    B --> C[ポリシーストア]
    D[Kubernetes] --> B
```

このサンプルアプリケーションは、以下のコンポーネントで構成されています：

1. **AuthZEN API サーバー**: AuthZEN仕様に準拠したRESTful APIを提供するサーバー
2. **ポリシーストア**: 認可ポリシーを管理するためのインメモリストア
3. **Kubernetes マニフェスト**: Kubernetes上にデプロイするためのマニフェストファイル

## 機能

このサンプルアプリケーションは、AuthZEN仕様で定義されている以下のAPIをサポートしています：

- **Access Evaluation API**: 単一の認可判断を行うAPI
- **Access Evaluations API**: 複数の認可判断を一度に行うAPI
- **Subject Search API**: 特定の条件に一致するSubjectを検索するAPI
- **Resource Search API**: 特定の条件に一致するResourceを検索するAPI
- **Action Search API**: 特定の条件に一致するActionを検索するAPI
- **メタデータディスカバリー**: PDPのメタデータを取得するためのAPI

## ディレクトリ構造

```
/
├── docs/                      # ドキュメント
│   ├── overview.md            # 概要ドキュメント
│   ├── api-spec.md            # API仕様ドキュメント
│   ├── information-model.md   # 情報モデルドキュメント
│   └── transport.md           # トランスポート層ドキュメント
├── kubernetes/                # Kubernetesマニフェスト
│   ├── deployment.yaml        # デプロイメント設定
│   └── service.yaml           # サービス設定
├── src/                       # ソースコード
│   ├── api/                   # APIの実装
│   │   ├── models.go          # データモデルの定義
│   │   └── server.go          # APIサーバーの実装
│   ├── policy/                # ポリシー関連の実装
│   │   └── store.go           # ポリシーストアの実装
│   ├── main.go                # メインエントリーポイント
│   ├── Dockerfile             # Dockerビルド設定
│   ├── deploy-to-kind.sh      # Kindへのデプロイスクリプト
│   ├── client-example.sh      # クライアント使用例スクリプト
│   └── README.md              # ソースコードのREADME
├── SPEC.md                    # 仕様適合性評価ドキュメント
└── README.md                  # このファイル
```

## 前提条件

- Go 1.18以上
- Docker
- Kind（Kubernetes in Docker）
- kubectl

## ビルドと実行

### ローカルでの実行

```bash
# ビルド
cd src
go build -o authzen-server

# 実行
./authzen-server
```

デフォルトでは、サーバーはポート8080でリッスンします。ポートを変更するには、`--port`フラグを使用します：

```bash
./authzen-server --port 9000
```

### Dockerイメージのビルド

```bash
cd src
docker build -t authzen-server:latest .
```

### Kindへのデプロイ

```bash
cd src
./deploy-to-kind.sh
```

このスクリプトは以下の処理を行います：
1. Kindクラスタの作成（まだ作成していない場合）
2. Dockerイメージのビルドとロード
3. Kubernetesマニフェストの適用

## APIの使用例

クライアント使用例スクリプトを使用して、APIをテストできます：

```bash
cd src
./client-example.sh
```

または、個別のAPIを直接呼び出すこともできます：

### Access Evaluation API

```bash
curl -X POST http://localhost:8080/access/v1/evaluation \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "type": "user",
      "id": "alice@example.com"
    },
    "resource": {
      "type": "document",
      "id": "123"
    },
    "action": {
      "name": "read"
    }
  }'
```

### メタデータディスカバリー

```bash
curl -X GET http://localhost:8080/.well-known/authzen-configuration
```

## 仕様適合性

このサンプルアプリケーションは、AuthZEN仕様に準拠しており、基本的な機能を正確に実装しています。詳細な仕様適合性評価については、[SPEC.md](SPEC.md)を参照してください。

## 実装の詳細

### ポリシーストア

このサンプルアプリケーションでは、シンプルなインメモリポリシーストアを使用しています。実際の本番環境では、データベースなどの永続的なストレージを使用することが推奨されます。

### 認可ロジック

認可ロジックは、以下のような単純なルールに基づいています：

1. Subject、Resource、Actionの組み合わせに一致するポリシーがある場合、そのポリシーの`Allow`値に基づいて判断します。
2. 一致するポリシーがない場合、デフォルトでは拒否（`false`）します。

### セキュリティ

このサンプルアプリケーションでは、簡単のために認証は実装していません。実際の本番環境では、OAuth 2.0などの認証メカニズムを実装することが推奨されます。

また、TLS（HTTPS）もデフォルトでは有効になっていませんが、`--tls`フラグと`--cert`、`--key`フラグを使用して有効にすることができます。

## 拡張と改善

このサンプルアプリケーションは、AuthZEN仕様の基本的な機能を示すために設計されています。実際の本番環境では、以下のような拡張や改善が考えられます：

1. **永続的なポリシーストレージ**: データベースなどを使用して、ポリシーを永続的に保存する
2. **より複雑な認可ロジック**: 属性ベースのアクセス制御（ABAC）や関係ベースのアクセス制御（ReBAC）などの高度な認可モデルをサポートする
3. **キャッシング**: パフォーマンスを向上させるために、認可判断の結果をキャッシュする
4. **監査ログ**: 認可判断の履歴を記録する
5. **管理API**: ポリシーの管理（追加、削除、更新）を行うためのAPIを提供する

## ライセンス

このサンプルアプリケーションは、MITライセンスの下で提供されています。
