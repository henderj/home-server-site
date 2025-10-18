set -e

git pull
docker build -t website:latest .
docker compose -f /srv/website/compose.yml up -d
