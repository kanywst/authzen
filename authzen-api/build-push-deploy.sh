#!/bin/bash
set -e

# Check if Docker Hub username is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <dockerhub-username>"
  echo "Example: $0 myusername"
  exit 1
fi

DOCKERHUB_USERNAME=$1
IMAGE_NAME="$DOCKERHUB_USERNAME/authzen-api:latest"

echo "Building Docker image: $IMAGE_NAME"
docker build -t $IMAGE_NAME .

echo "Pushing Docker image to Docker Hub"
echo "Please make sure you're logged in to Docker Hub (docker login)"
docker push $IMAGE_NAME

echo "Updating Kubernetes deployment file"
sed "s/\${DOCKERHUB_USERNAME}/$DOCKERHUB_USERNAME/g" kubernetes/deployment.yaml > kubernetes/deployment-updated.yaml

echo "Applying Kubernetes resources"
kubectl apply -f kubernetes/service.yaml
kubectl apply -f kubernetes/deployment-updated.yaml

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
