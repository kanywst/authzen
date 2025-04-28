# AuthZEN トランスポート詳細

このドキュメントでは、AuthZEN（Authorization API）のトランスポート層について詳細に説明します。

## トランスポートの概要

AuthZENは、PDPとPEP間の通信のためのトランスポート層を定義しています。現在の仕様では、HTTPSバインディングが必須として定義されていますが、将来的には他のトランスポート（例：gRPC）もプロファイルの形で定義される可能性があります。

```mermaid
graph LR
    A[PEP: Policy Enforcement Point] <--> B[HTTPS]
    B <--> C[PDP: Policy Decision Point]
```

## HTTPSバインディング

HTTPSバインディングは、AuthZEN仕様に準拠するPDPによって実装されなければなりません。

### エンドポイント

AuthZEN APIの各エンドポイントは、バージョン番号を含むパスで提供されます。バージョン1.0のAPIでは、すべてのエンドポイントは `/v1/` で始まります。

主要なエンドポイントは以下の通りです：

- Access Evaluation API: `/access/v1/evaluation`
- Access Evaluations API: `/access/v1/evaluations`
- Subject Search API: `/access/v1/search/subject`
- Resource Search API: `/access/v1/search/resource`
- Action Search API: `/access/v1/search/action`

### リクエストとレスポンスの形式

すべてのリクエストとレスポンスは、`application/json`コンテンツタイプを使用します。

### HTTPSヘッダー

#### 認証

AuthZEN APIへのアクセスは、通常、認証が必要です。認証方法は、この仕様の範囲外ですが、一般的には以下のような方法が使用されます：

- OAuth 2.0トークンを使用した`Authorization: Bearer <token>`ヘッダー
- APIキーを使用した`Authorization: ApiKey <key>`ヘッダー
- 基本認証を使用した`Authorization: Basic <credentials>`ヘッダー

例：
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### リクエスト識別

すべてのリクエストには、一意に識別するためのリクエスト識別子を含めることができます。リクエスト識別子は、`X-Request-ID`ヘッダーを使用して提供されます。

例：
```
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716
```

PDPは、リクエストに`X-Request-ID`ヘッダーが含まれている場合、レスポンスにも同じ識別子を含める必要があります。

### HTTPSステータスコード

AuthZEN APIは、以下のHTTPSステータスコードを使用します：

| コード | 説明 | 使用例 |
|------|--------------|-------------------|
| 200  | OK | 認可リクエストが正常に処理された場合（許可/拒否に関わらず） |
| 400  | Bad Request | リクエストの形式が不正な場合 |
| 401  | Unauthorized | 認証情報が不足または無効な場合 |
| 403  | Forbidden | 認証されたが、APIへのアクセス権限がない場合 |
| 500  | Internal Server Error | サーバー内部でエラーが発生した場合 |

**注意**: 認可判断の結果（許可/拒否）は、HTTPSステータスコードではなく、レスポンスボディの`decision`フィールドで示されます。認可リクエストが拒否された場合でも、リクエスト自体が正常に処理されれば、ステータスコードは`200 OK`になります。

## Access Evaluation API

### リクエスト

```
POST /access/v1/evaluation HTTP/1.1
Host: pdp.mycompany.com
Authorization: Bearer <myoauthtoken>
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "subject": {
    "type": "user",
    "id": "alice@acmecorp.com"
  },
  "resource": {
    "type": "document",
    "id": "123"
  },
  "action": {
    "name": "read"
  },
  "context": {
    "time": "2024-10-26T01:22-07:00"
  }
}
```

### レスポンス

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "decision": true
}
```

または、拒否の場合：

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "decision": false,
  "context": {
    "reason": "リソースへのアクセス権限がありません"
  }
}
```

## Access Evaluations API

### リクエスト

```
POST /access/v1/evaluations HTTP/1.1
Host: pdp.mycompany.com
Authorization: Bearer <myoauthtoken>
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "subject": {
    "type": "user",
    "id": "alice@acmecorp.com"
  },
  "context": {
    "time": "2024-05-31T15:22-07:00"
  },
  "action": {
    "name": "read"
  },
  "evaluations": [
    {
      "resource": {
        "type": "document",
        "id": "123"
      }
    },
    {
      "resource": {
        "type": "document",
        "id": "456"
      }
    },
    {
      "action": {
        "name": "write"
      },
      "resource": {
        "type": "document",
        "id": "789"
      }
    }
  ]
}
```

### レスポンス

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "evaluations": [
    {
      "decision": true
    },
    {
      "decision": false,
      "context": {
        "reason": "リソースが見つかりません"
      }
    },
    {
      "decision": false,
      "context": {
        "reason": "書き込み権限がありません"
      }
    }
  ]
}
```

## Subject Search API

### リクエスト

```
POST /access/v1/search/subject HTTP/1.1
Host: pdp.mycompany.com
Authorization: Bearer <myoauthtoken>
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "subject": {
    "type": "user"
  },
  "action": {
    "name": "read"
  },
  "resource": {
    "type": "document",
    "id": "123"
  }
}
```

### レスポンス

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "results": [
    {
      "type": "user",
      "id": "alice@acmecorp.com"
    },
    {
      "type": "user",
      "id": "bob@acmecorp.com"
    }
  ],
  "page": {
    "next_token": "alsehrq3495u8"
  }
}
```

## Resource Search API

### リクエスト

```
POST /access/v1/search/resource HTTP/1.1
Host: pdp.mycompany.com
Authorization: Bearer <myoauthtoken>
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "subject": {
    "type": "user",
    "id": "alice@acmecorp.com"
  },
  "action": {
    "name": "read"
  },
  "resource": {
    "type": "document"
  }
}
```

### レスポンス

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "results": [
    {
      "type": "document",
      "id": "123"
    },
    {
      "type": "document",
      "id": "456"
    }
  ],
  "page": {
    "next_token": "alsehrq3495u8"
  }
}
```

## Action Search API

### リクエスト

```
POST /access/v1/search/action HTTP/1.1
Host: pdp.mycompany.com
Authorization: Bearer <myoauthtoken>
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "subject": {
    "type": "user",
    "id": "alice@acmecorp.com"
  },
  "resource": {
    "type": "document",
    "id": "123"
  }
}
```

### レスポンス

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: bfe9eb29-ab87-4ca3-be83-a1d5d8305716

{
  "results": [
    {
      "name": "read"
    },
    {
      "name": "write"
    }
  ],
  "page": {
    "next_token": ""
  }
}
```

## メタデータディスカバリー

AuthZENは、PDPのメタデータを取得するためのエンドポイントを定義しています。このエンドポイントは、Well-Known URIの仕組みを使用しています。

### リクエスト

```
GET /.well-known/authzen-configuration HTTP/1.1
Host: pdp.mycompany.com
```

### レスポンス

```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "policy_decision_point": "https://pdp.mycompany.com",
  "access_evaluation_endpoint": "https://pdp.mycompany.com/access/v1/evaluation",
  "access_evaluations_endpoint": "https://pdp.mycompany.com/access/v1/evaluations",
  "search_subject_endpoint": "https://pdp.mycompany.com/access/v1/search/subject",
  "search_resource_endpoint": "https://pdp.mycompany.com/access/v1/search/resource",
  "search_action_endpoint": "https://pdp.mycompany.com/access/v1/search/action"
}
```

## セキュリティ考慮事項

### 通信の整合性と機密性

PEPとPDP間の通信は、TLSを使用して保護する必要があります。TLSの設定は、以下の要件を満たす必要があります：

- TLS 1.2以上を使用
- 強力な暗号スイートを使用
- 適切な証明書検証を実施

### ポリシーの機密性と送信者認証

PDPは、呼び出し元のPEPを認証する必要があります。認証方法としては、以下のような選択肢があります：

- OAuth 2.0トークン
- 相互TLS（mTLS）
- APIキー

### 可用性とDoS対策

PDPは、以下のような攻撃に対する保護を実施する必要があります：

- リクエストペイロードのサイズ制限
- リクエスト数の制限（レート制限）
- 無効なJSONの検出
- ネストされたJSONの攻撃の検出
- メモリ消費の制限

## 実装例

このリポジトリには、AuthZEN仕様に準拠したHTTPSバインディングの実装例が含まれています。詳細は[サンプルアプリケーション](../src/README.md)を参照してください。
