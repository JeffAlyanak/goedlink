#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

apt-get update
apt-get install -y jq

# Get the repository and release ID from the event payload
REPO_OWNER=$(jq -r '.repository.owner.login' "$GITHUB_EVENT_PATH")
REPO_NAME=$(jq -r '.repository.name' "$GITHUB_EVENT_PATH")
RELEASE_ID=$(jq -r '.release.id' "$GITHUB_EVENT_PATH")
ASSET_NAME=$1
FILE_PATH=$2

echo "https://forge.rights.ninja/api/v1/repos/${REPO_OWNER}/${REPO_NAME}/releases/${RELEASE_ID}/assets?name=${ASSET_NAME}&token=${GITHUB_TOKEN}"

# Upload the release asset
curl -X POST \
  "https://forge.rights.ninja/api/v1/repos/${REPO_OWNER}/${REPO_NAME}/releases/${RELEASE_ID}/assets?name=${ASSET_NAME}&token=${GITHUB_TOKEN}" \
  -H "accept: application/json" \
  -H "Content-Type: multipart/form-data" \
  -F "attachment=@${FILE_PATH}"
