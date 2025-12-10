#!/usr/bin/env bash
set -e
cd /opt/cloudcop
git reset --hard HEAD~1
docker compose -f infra/docker-compose.yml up -d