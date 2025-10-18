set -e

# git pull
docker build -t website:latest .
# currently, the nixos config is setup to restart this service when it's stopped.
# so, to restart the service with the new image, we just stop it and let nixos restart it.
docker compose -f /srv/website/compose.yml up -d
