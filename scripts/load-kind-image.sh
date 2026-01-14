#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="rp-practice"

echo "Fetching Docker images matching patterns: user:latest, order:latest, payment:latest, notification:latest, inventory:latest..."

# Получаем список образов, заканчивающихся на ':latest', и начинающихся с одного из указанных префиксов
IMAGES=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep -E '^(rp-user|rp-order|rp-payment|rp-notification|rp-inventory):latest$')

if [ -z "$IMAGES" ]; then
  echo "No images found matching the pattern (user|order|payment|notification|inventory):latest." >&2
  exit 1
fi

echo "Found images:"
echo "$IMAGES"
echo "------------------------"

for IMAGE in $IMAGES; do
  echo "Loading image: $IMAGE to kind cluster: $CLUSTER_NAME"
  # Выполняем kind load docker-image
  kind load docker-image "$IMAGE" --name "$CLUSTER_NAME"
done

echo "All matching images have been loaded into the kind cluster '$CLUSTER_NAME'."