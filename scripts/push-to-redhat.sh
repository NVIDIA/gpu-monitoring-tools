#!/bin/bash

IMAGE_TAG=${CIRCLE_TAG/v/''}
IMAGE=gcr.io/$GCLOUD_PROJECT/$IMAGE_ID:$IMAGE_TAG

# Kepp manifest digest of the container image
docker images -q --filter reference=$IMAGE > digest.txt

echo $REDHAT_REGISTRY_KEY | docker login -u unused scan.connect.redhat.com --password-stdin
docker tag $IMAGE scan.connect.redhat.com/$REDHAT_PROJECT_ID/$IMAGE_ID:$IMAGE_TAG
docker push scan.connect.redhat.com/$REDHAT_PROJECT_ID/$IMAGE_ID:$IMAGE_TAG