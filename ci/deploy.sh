#!/bin/bash
# Deploy the helloworld service to Kubernetes.
#
# Pushes the multi-arch image index, derives its immutable @sha256 digest from
# the built artifact, generates a Kustomize overlay pinned to that digest, and
# applies it. Deploying by digest (not a mutable tag) makes rollouts reproducible.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

PUSH_TARGET="//services/helloworld:push"
IMAGE_INDEX="//services/helloworld:image_index"
IMAGE_REPO="ghcr.io/esurdam/go-grpc-bazel-example/services/helloworld"
OVERLAY_DIR="deploy/helloworld/overlays/live"

command -v kubectl >/dev/null || { echo "deploy: kubectl not found on PATH" >&2; exit 1; }
command -v jq >/dev/null || { echo "deploy: jq not found on PATH" >&2; exit 1; }

# 1. Build the multi-arch index and derive its digest from the OCI layout.
#    This digest is content-addressed, so it is identical to what gets pushed.
bazel build "$IMAGE_INDEX"
INDEX_DIR="$(bazel cquery --output=files "$IMAGE_INDEX" 2>/dev/null | head -1)"
DIGEST="$(jq -r '.manifests[0].digest' "$INDEX_DIR/index.json")"
if [ -z "$DIGEST" ] || [ "$DIGEST" = "null" ]; then
  echo "deploy: could not read image digest from $INDEX_DIR/index.json" >&2
  exit 1
fi

# 2. Push the multi-arch index (SHA-tagged via :stamped).
bazel run "$PUSH_TARGET"

# 3. Generate the (gitignored) overlay pinned to the digest, with the git commit
#    recorded as an annotation for traceability.
GIT_SHA="$(git rev-parse --short HEAD)"
mkdir -p "$OVERLAY_DIR"
cat > "$OVERLAY_DIR/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

images:
  - name: ${IMAGE_REPO}
    digest: ${DIGEST}

commonAnnotations:
  app.kubernetes.io/version: "${GIT_SHA}"
EOF

# 4. Render as a self-check (fails fast if the overlay is invalid), then apply.
echo "deploy: rendering ${IMAGE_REPO}@${DIGEST} (commit ${GIT_SHA})"
kubectl kustomize "$OVERLAY_DIR"
kubectl apply -k "$OVERLAY_DIR"
