#!/bin/bash

# Exit on error
set -e

# Default API endpoint
API_URL=${1:-"http://localhost:8080"}

# Check if jq command exists
if ! command -v jq &> /dev/null; then
    echo "jq command not found. Please install it."
    exit 1
fi

echo "AuthZEN API Client Example"
echo "API URL: $API_URL"
echo

# Get metadata
echo "=== Get Metadata ==="
curl -s -X GET "$API_URL/.well-known/authzen-configuration" | jq
echo

# Authorization request (allow)
echo "=== Authorization Request (Allow) ==="
curl -s -X POST "$API_URL/access/v1/evaluation" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "type": "user",
      "id": "user:alice"
    },
    "resource": {
      "type": "document",
      "id": "document:123"
    },
    "action": {
      "name": "read"
    }
  }' | jq
echo

# Authorization request (deny)
echo "=== Authorization Request (Deny) ==="
curl -s -X POST "$API_URL/access/v1/evaluation" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "type": "user",
      "id": "user:charlie"
    },
    "resource": {
      "type": "document",
      "id": "document:123"
    },
    "action": {
      "name": "read"
    }
  }' | jq
echo

# Multiple authorization requests
echo "=== Multiple Authorization Requests ==="
curl -s -X POST "$API_URL/access/v1/evaluations" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "type": "user",
      "id": "user:alice"
    },
    "evaluations": [
      {
        "resource": {
          "type": "document",
          "id": "document:123"
        },
        "action": {
          "name": "read"
        }
      },
      {
        "resource": {
          "type": "document",
          "id": "document:123"
        },
        "action": {
          "name": "write"
        }
      }
    ]
  }' | jq
echo

# Subject search
echo "=== Subject Search ==="
curl -s -X POST "$API_URL/access/v1/search/subject" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "type": "user"
    },
    "resource": {
      "type": "document",
      "id": "document:123"
    },
    "action": {
      "name": "read"
    }
  }' | jq
echo

# Resource search
echo "=== Resource Search ==="
curl -s -X POST "$API_URL/access/v1/search/resource" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "type": "user",
      "id": "user:alice"
    },
    "resource": {
      "type": "document"
    },
    "action": {
      "name": "read"
    }
  }' | jq
echo

# Action search
echo "=== Action Search ==="
curl -s -X POST "$API_URL/access/v1/search/action" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "type": "user",
      "id": "user:alice"
    },
    "resource": {
      "type": "document",
      "id": "document:123"
    }
  }' | jq
echo

# List policies
echo "=== List Policies ==="
curl -s -X GET "$API_URL/v1/policies" | jq
echo

echo "All API requests completed."
