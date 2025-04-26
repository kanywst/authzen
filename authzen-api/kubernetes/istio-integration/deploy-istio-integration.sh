#!/bin/bash
set -e

# 必要なツールの確認
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed. Aborting."; exit 1; }
command -v istioctl >/dev/null 2>&1 || { echo "istioctl is required but not installed. Aborting."; exit 1; }
command -v kind >/dev/null 2>&1 || { echo "kind is required but not installed. Aborting."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "docker is required but not installed. Aborting."; exit 1; }

# クラスタが存在するか確認
if ! kind get clusters | grep -q "authzen-sample"; then
  echo "Creating Kind cluster: authzen-sample"
  kind create cluster --name authzen-sample
else
  echo "Kind cluster authzen-sample already exists"
fi

# Istioのインストール
echo "Installing Istio..."
istioctl install --set profile=demo -y

# 名前空間にIstio自動インジェクションラベルを設定
kubectl label namespace default istio.io/inject=enabled --overwrite

# AuthZEN APIのビルドとデプロイ
echo "Building and deploying AuthZEN API..."
cd ../..
docker build -t authzen-api:latest .
kind load docker-image authzen-api:latest --name authzen-sample
kubectl apply -f kubernetes/deployment.yaml
kubectl apply -f kubernetes/service.yaml

# サンプルアプリケーションのデプロイ
echo "Deploying sample application..."
kubectl apply -f kubernetes/istio-integration/sample-app.yaml

# Istio Ingressゲートウェイのデプロイ
echo "Deploying Istio Ingress Gateway..."
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: ingressgateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
EOF

# Istio認可設定のデプロイ
echo "Deploying Istio Authorization configuration..."
kubectl apply -f kubernetes/istio-integration/extension-provider.yaml
kubectl apply -f kubernetes/istio-integration/authzpolicy.yaml

# デプロイメントの状態を確認
echo "Waiting for deployments to be ready..."
kubectl rollout status deployment/authzen-api
kubectl rollout status deployment/sample-app

echo "Deployment completed successfully!"
echo ""
echo "To test the integration, run:"
echo "  kubectl port-forward -n istio-system svc/istio-ingressgateway 8000:80"
echo ""
echo "Then in another terminal:"
echo 'curl -H "x-user-id: user:alice" http://localhost:8000/sample'
echo 'curl -H "x-user-id: user:bob" http://localhost:8000/sample'
echo 'curl -H "x-user-id: user:alice" http://localhost:8000/sample/admin'
echo 'curl -H "x-user-id: user:bob" http://localhost:8000/sample/admin'
