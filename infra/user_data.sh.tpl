#!/bin/bash
set -euo pipefail

dnf update -y
dnf install -y docker
systemctl enable --now docker

aws ecr get-login-password --region ${region} \
  | docker login --username AWS --password-stdin "$(echo "${ecr_url}" | cut -d'/' -f1)"

docker pull "${ecr_url}:${image_tag}"

docker rm -f app 2>/dev/null || true

docker run -d --name app --restart unless-stopped -p 8080:8080 \
  -e PORT=8080 \
  -e REDIS_ADDR="${redis_addr}" \
  -e AUTH_USER="${auth_user}" \
  -e AUTH_PASS="${auth_pass}" \
  -e CACHE_TTL_SECONDS="${cache_ttl_seconds}" \
  "${ecr_url}:${image_tag}"
