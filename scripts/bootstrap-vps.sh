#!/usr/bin/env bash
set -e
apt update && apt install -y docker.io docker-compose git
useradd -m deploy || true
su - deploy -c "git clone https://github.com/PhantomInTheWire/CloudCop.git /opt/cloudcop"
cd /opt/cloudcop/infra
docker compose up -d