#!/bin/bash
set -e

echo -e "\nStarting CI stage: ${CI_JOB_STAGE}"

PROJECT_TYPE="portal"
PROJECT_DIR="/var/www/sie/${PROJECT_TYPE}/${CI_PROJECT_NAME}"
REPO_URL="https://${GITLAB_PAT}@gitlab.iimlab.id/${CI_PROJECT_PATH}.git"
CI_PROJECT_NAME_LOWER="$(echo "${CI_PROJECT_NAME}" | tr '[:upper:]' '[:lower:]')"

export CI_PROJECT_NAME_LOWER

pull_code() {
  echo -e "\n[Pull] Pulling code from repository..."
  cd /var/www/
  if [ ! -d "${CI_PROJECT_NAME}" ]; then
    echo "Cloning project in $(pwd) ..."
    git clone "${REPO_URL}" || { echo "Clone failed"; exit 1; }
  fi
  cd "${PROJECT_DIR}"
  git reset --hard
  git pull origin main
}

build_image() {
  echo -e "\n[Build] Building Docker image..."
  cd "${PROJECT_DIR}"
  VERSION=$(date +%Y%m%d%H%M%S)
  IMAGE_NAME="${CI_PROJECT_NAME_LOWER}:${VERSION}"
  REGISTRY_IMAGE="localhost:8001/${CI_PROJECT_NAME_LOWER}:${VERSION}"

    if ! docker build -t ${IMAGE_NAME} .; then
        echo "Docker build failed"
        exit 1
    fi

    if ! docker tag ${IMAGE_NAME} ${REGISTRY_IMAGE}; then
        echo "Docker tag for registry failed"
        exit 1
    fi

    if ! docker push ${REGISTRY_IMAGE}; then
        echo "Docker push to registry failed"
        exit 1
    fi
    if ! docker tag ${IMAGE_NAME} localhost:8001/${CI_PROJECT_NAME_LOWER}:latest; then
        echo "Docker tag as latest failed"
        exit 1
    fi

    if ! docker push localhost:8001/${CI_PROJECT_NAME_LOWER}:latest; then
        echo "Docker push latest failed"
        exit 1
    fi

    # Save version for deployment stage
    echo "${VERSION}" > .image_version
    echo "Built and pushed image version: ${VERSION}"

    docker rmi ${IMAGE_NAME} || true
    docker rmi localhost:8001/${CI_PROJECT_NAME_LOWER}:${VERSION} || true
}

deploy_app() {
  echo -e "\n[Deploy] Deploying application..."
  cd "${PROJECT_DIR}"

  # Try to read version from build stage artifact
  if [ -f ".image_version" ]; then
    VERSION=$(cat .image_version)
    echo "Using version from build stage: ${VERSION}"
  else
    # Fallback: try to get the latest image version from registry
    VERSION=$(docker images localhost:8001/${CI_PROJECT_NAME_LOWER} --format "{{.Tag}}" | grep -v "^latest$" | grep -v "^<none>$" | sort -r | head -n 1)
    
    # If still no version found, pull and use latest
    if [ -z "$VERSION" ] || [ "$VERSION" = "<none>" ]; then
      echo "No versioned image found, pulling and using latest"
      docker pull localhost:8001/${CI_PROJECT_NAME_LOWER}:latest || true
      VERSION="latest"
    fi
  fi

  export IMAGE_TAG=$VERSION
  export CI_REGISTRY_IMAGE="localhost:8001/${CI_PROJECT_NAME_LOWER}"
  
  echo "Deploying with IMAGE_TAG=${IMAGE_TAG}"
  echo "Registry: ${CI_REGISTRY_IMAGE}"

  # Stop existing containers
  echo "Stopping existing containers..."
  if command -v docker-compose &> /dev/null; then
    docker-compose down || true
  elif docker compose version &> /dev/null; then
    docker compose down || true
  fi

  # Pull latest images
  echo "Pulling latest images..."
  if command -v docker-compose &> /dev/null; then
    docker-compose pull || true
  elif docker compose version &> /dev/null; then
    docker compose pull || true
  fi

  # Start services
  echo "Starting services..."
  if command -v docker-compose &> /dev/null; then
    docker-compose up -d
  elif docker compose version &> /dev/null; then
    docker compose up -d
  fi

  echo "Deployment completed successfully!"
}

case "$CI_JOB_STAGE" in
  pull) pull_code ;;
  build) build_image ;;
  deploy) deploy_app ;;
  *) echo "Unknown stage: $CI_JOB_STAGE"; exit 1 ;;
esac
