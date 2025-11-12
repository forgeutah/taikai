#!/bin/bash
# Push Taikai Docker images to registry

set -e

VERSION=${1:-latest}
REGISTRY=${DOCKER_REGISTRY:-forgeutah}

echo "Pushing Taikai Docker images..."
echo "Version: $VERSION"
echo "Registry: $REGISTRY"

# Push all images
echo "Pushing server image..."
docker push ${REGISTRY}/taikai:${VERSION}
docker push ${REGISTRY}/taikai:latest

echo "Pushing worker image..."
docker push ${REGISTRY}/taikai-worker:${VERSION}
docker push ${REGISTRY}/taikai-worker:latest

echo "Pushing migrate image..."
docker push ${REGISTRY}/taikai-migrate:${VERSION}
docker push ${REGISTRY}/taikai-migrate:latest

echo ""
echo "✅ All images pushed successfully!"
echo ""
echo "Images available:"
echo "  ${REGISTRY}/taikai:${VERSION}"
echo "  ${REGISTRY}/taikai-worker:${VERSION}"
echo "  ${REGISTRY}/taikai-migrate:${VERSION}"
