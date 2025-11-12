#!/bin/bash
# Build script for Taikai Docker images

set -e

VERSION=${1:-latest}
REGISTRY=${DOCKER_REGISTRY:-forgeutah}

echo "Building Taikai Docker images..."
echo "Version: $VERSION"
echo "Registry: $REGISTRY"

# Build multi-stage Dockerfile
echo "Building server image..."
docker build \
  -f Dockerfile.prod \
  --target server \
  -t ${REGISTRY}/taikai:${VERSION} \
  -t ${REGISTRY}/taikai:latest \
  .

echo "Building worker image..."
docker build \
  -f Dockerfile.prod \
  --target worker \
  -t ${REGISTRY}/taikai-worker:${VERSION} \
  -t ${REGISTRY}/taikai-worker:latest \
  .

echo "Building migrate image..."
docker build \
  -f Dockerfile.prod \
  --target migrate \
  -t ${REGISTRY}/taikai-migrate:${VERSION} \
  -t ${REGISTRY}/taikai-migrate:latest \
  .

echo ""
echo "✅ Images built successfully!"
echo ""
echo "To push images:"
echo "  docker push ${REGISTRY}/taikai:${VERSION}"
echo "  docker push ${REGISTRY}/taikai-worker:${VERSION}"
echo "  docker push ${REGISTRY}/taikai-migrate:${VERSION}"
echo ""
echo "Or push all at once:"
echo "  ./push-docker.sh $VERSION"
