set -euf -o pipefail
docker buildx build --platform linux/amd64 --provenance=false -t heartbeat:latest .