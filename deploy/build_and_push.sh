git fetch --tags
LATEST_TAG=$(git tag -l | grep "$1" | sort -V | tail -n 1 | cut -d' ' -f1)
BUILD_TIME=$(date -u +"%Y-%m-%d_%H:%M_%Z")
echo "Building version: $LATEST_TAG"
echo "Build time: $BUILD_TIME"
docker build --build-arg GIT_TAG=${LATEST_TAG} --build-arg BUILD_TIME=${BUILD_TIME} --platform linux/amd64 -f deploy/Dockerfile -t feesh-api . &&
docker tag feesh docker.io/1f47e/feesh-api:latest &&
docker push docker.io/1f47e/feesh-api:latest


