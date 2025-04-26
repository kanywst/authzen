#!/bin/bash
set -e

# APIサーバーのエンドポイント
API_ENDPOINT="http://localhost:8080"

# 認可リクエストの送信
function check_authorization() {
  local principal=$1
  local resource=$2
  local action=$3

  echo "Checking authorization: $principal -> $action -> $resource"
  
  response=$(curl -s -X POST "$API_ENDPOINT/v1/authorize" \
    -H "Content-Type: application/json" \
    -d "{
      \"principal\": {\"id\": \"$principal\"},
      \"resource\": {\"id\": \"$resource\"},
      \"action\": \"$action\"
    }")
  
  decision=$(echo $response | jq -r '.decision')
  reason=$(echo $response | jq -r '.reason // "N/A"')
  
  echo "Decision: $decision"
  echo "Reason: $reason"
  echo "---"
}

# ポリシーの一覧表示
function list_policies() {
  echo "Listing all policies:"
  curl -s -X GET "$API_ENDPOINT/v1/policies" | jq .
  echo "---"
}

# サーバーが起動しているか確認
if ! curl -s "$API_ENDPOINT/health" > /dev/null; then
  echo "Error: Authorization API server is not running."
  echo "Please start the server with: kubectl port-forward svc/authzen-api 8080:8080"
  exit 1
fi

# テストケースの実行
echo "=== Authorization API Client Example ==="
echo ""

list_policies

echo "Testing authorization requests:"
check_authorization "user:alice" "document:report" "read"   # 許可されるはず
check_authorization "user:alice" "document:report" "write"  # 許可されるはず
check_authorization "user:bob" "document:report" "read"     # 許可されるはず
check_authorization "user:bob" "document:report" "write"    # 拒否されるはず
check_authorization "user:charlie" "document:report" "read" # 拒否されるはず
check_authorization "user:dave" "document:report" "read"    # ポリシーがないため拒否されるはず

echo "Client example completed."
