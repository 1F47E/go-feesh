docker build --platform linux/amd64 -f deploy/Dockerfile -t feesh-api . &&
docker tag feesh docker.io/1f47e/feesh-api:latest &&
docker push docker.io/1f47e/feesh-api:latest


