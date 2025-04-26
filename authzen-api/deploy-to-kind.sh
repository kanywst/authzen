#!/bin/bash
set -e

# クラスタが存在するか確認
if ! kind get clusters | grep -q "authzen-sample"; then
  echo "Creating Kind cluster: authzen-sample"
  kind create cluster --name authzen-sample
else
  echo "Kind cluster authzen-sample already exists"
fi

# Dockerイメージのビルド
echo "Building Docker image: authzen-api:latest"
docker build -t authzen-api:latest .

# イメージをKindにロード
echo "Loading image into Kind cluster"
kind load docker-image authzen-api:latest --name authzen-sample

# Kubernetesリソースの適用
echo "Applying Kubernetes resources"
kubectl apply -f kubernetes/

# デプロイメントの状態を確認
echo "Waiting for deployment to be ready"
kubectl rollout status deployment/authzen-api

echo "Deployment completed successfully!"
echo ""
echo "To test the API, run:"
echo "  kubectl port-forward svc/authzen-api 8080:8080"
echo ""
echo "Then in another terminal:"
echo 'curl -X POST http://localhost:8080/v1/authorize \'
echo '  -H "Content-Type: application/json" \'
echo '  -d '\''{"principal": {"id": "user:alice"}, "resource": {"id": "document:report"}, "action": "read"}'\'''
