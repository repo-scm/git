#!/bin/bash

echo "Fetching latest fuse-overlayfs release..."
FUSE_OVERLAYFS_VERSION=$(curl -s "https://api.github.com/repos/containers/fuse-overlayfs/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$FUSE_OVERLAYFS_VERSION" ]; then
    echo "Failed to fetch latest release version"
    exit 1
fi

FUSE_OVERLAYFS_URL="https://github.com/containers/fuse-overlayfs/releases/download/${FUSE_OVERLAYFS_VERSION}/fuse-overlayfs-x86_64"

EMBED_DIR="embedded"
BINARY_PATH="${EMBED_DIR}/fuse-overlayfs"

echo "Downloading fuse-overlayfs ${FUSE_OVERLAYFS_VERSION} for embedding..."

mkdir -p "${EMBED_DIR}"

if curl -L -f -o "${BINARY_PATH}" "${FUSE_OVERLAYFS_URL}"; then
    chmod +x "${BINARY_PATH}"
    echo "Successfully downloaded fuse-overlayfs to ${BINARY_PATH}"
else
    echo "Failed to download fuse-overlayfs from ${FUSE_OVERLAYFS_URL}"
    exit 1
fi
