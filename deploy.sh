#!/usr/bin/env bash
set -euo pipefail

REGION=$(terraform -chdir=infra output -raw aws_region)
ECR_URL=$(terraform -chdir=infra output -raw ecr_repository_url)
ASG_NAME=$(terraform -chdir=infra output -raw autoscaling_group_name)
IMAGE_TAG=${IMAGE_TAG:-latest}

echo "==> Building image for linux/amd64..."
docker build --platform linux/amd64 -t "${ECR_URL}:${IMAGE_TAG}" .

echo "==> Logging in to ECR..."
aws ecr get-login-password --region "${REGION}" \
  | docker login --username AWS --password-stdin "${ECR_URL%%/*}"

echo "==> Pushing ${ECR_URL}:${IMAGE_TAG}..."
docker push "${ECR_URL}:${IMAGE_TAG}"

echo "==> Triggering rolling instance refresh on ${ASG_NAME}..."
REFRESH_ID=$(aws autoscaling start-instance-refresh \
  --region "${REGION}" \
  --auto-scaling-group-name "${ASG_NAME}" \
  --preferences '{"MinHealthyPercentage": 50, "InstanceWarmup": 90}' \
  --output text --query 'InstanceRefreshId')

echo "==> Waiting for instance refresh ${REFRESH_ID} to complete..."
while true; do
  STATUS=$(aws autoscaling describe-instance-refreshes \
    --region "${REGION}" \
    --auto-scaling-group-name "${ASG_NAME}" \
    --instance-refresh-ids "${REFRESH_ID}" \
    --output text --query 'InstanceRefreshes[0].Status')
  echo "    status: ${STATUS}"
  case "${STATUS}" in
    Successful) break ;;
    Failed|Cancelled) echo "Instance refresh ${STATUS}"; exit 1 ;;
  esac
  sleep 15
done

ALB=$(terraform -chdir=infra output -raw alb_dns_name)
echo ""
echo "Deploy complete. Service URL: ${ALB}"
