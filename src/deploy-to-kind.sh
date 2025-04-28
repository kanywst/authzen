#!/bin/bash

# Exit on error
set -e

# Get current directory
CURRENT_DIR=$(pwd)
SCRIPT_DIR=$(dirname "$0")
cd "$SCRIPT_DIR"

echo "Deploying AuthZEN API server to Kind..."

# Check if Kind cluster exists
if ! kind get clusters | grep -q "authzen"; then
    echo "Creating Kind cluster 'authzen'..."
    kind create cluster --name authzen
else
    echo "Kind cluster 'authzen' already exists."
fi

# Build Docker image
echo "Building Docker image..."
docker build -t authzen-server:latest .

# Load Docker image into Kind cluster
echo "Loading Docker image into Kind cluster..."
kind load docker-image authzen-server:latest --name authzen

# Apply Kubernetes manifests
echo "Applying Kubernetes manifests..."
kubectl apply -f ../kubernetes/

# Verify deployment
echo "Verifying deployment status..."
kubectl get pods -l app=authzen-api
kubectl get services -l app=authzen-api

echo "Deployment complete. You can access the API with the following commands:"
echo "kubectl port-forward svc/authzen-api 8080:8080"
echo "curl -X GET http://localhost:8080/.well-known/authzen-configuration"

# Return to original directory
cd "$CURRENT_DIR"
