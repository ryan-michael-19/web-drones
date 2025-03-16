# TODO: Migrate to docker compose?
set -euf -o pipefail

ACCOUNT_ID="${1}"
REPOSITORY_URI="$ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com"
REPOSITORY_NAME="${REPOSITORY_URI}/heartbeat"

# build
docker buildx build --platform linux/amd64 --provenance=false -t heartbeat:latest -f heartbeat.Dockerfile .

# deploy
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $REPOSITORY_URI
docker tag heartbeat:latest "${REPOSITORY_NAME}:latest"
docker push ${REPOSITORY_NAME}

aws lambda update-function-code --function-name heartbeat --image-uri $REPOSITORY_NAME:latest --publish