#!/usr/bin/env bash

SCRIPT_DIR=$(cd $(dirname $0); pwd)

PUSH=$1
DATE="$(date "XXXXXXXX")"
REPOSITORY_NAME="latonaio"
IMAGE_NAME="mysql-backup-to-usb-storage"
DOCKERFILE_DIR=${SCRIPT_DIR}/../
DOCKERFILE_NAME="Dockerfile"

# build 
DOCKER_BUILDKIT=1 docker build --secret id=ssh,src=$HOME/.ssh/id_rsa -f ${DOCKERFILE_DIR}/${DOCKERFILE_NAME} -t ${REPOSITORY_NAME}/${IMAGE_NAME}:"${DATE}" .
docker tag ${REPOSITORY_NAME}/${IMAGE_NAME}:"${DATE}" ${REPOSITORY_NAME}/${IMAGE_NAME}:latest

if [[ $PUSH == "push" ]]; then
    docker push ${REPOSITORY_NAME}/${IMAGE_NAME}:"${DATE}"
    docker push ${REPOSITORY_NAME}/${IMAGE_NAME}:latest
fi
