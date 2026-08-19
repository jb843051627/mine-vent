#!/bin/bash
set -e

NAME="${1:?usage: build_benzhi_docker.sh <image-name> <platform>}"
PLATFORM="${2:?usage: build_benzhi_docker.sh <image-name> <platform>}"

IMAGE="benzhi/${NAME}:latest"

echo "Building ${IMAGE} for ${PLATFORM}..."

docker build \
  --platform "${PLATFORM}" \
  -f benzhi.Dockerfile \
  -t "${IMAGE}" \
  .

echo "Done: ${IMAGE} (${PLATFORM})"
