#!/bin/bash

set -e

# Pull the turtle docker image.
docker pull desertbit/turtle

# Install the files.
cp -f ./scripts/turtle.service /etc/systemd/system/turtle.service
cp -f ./scripts/host/turtle /bin/turtle
cp -f ./scripts/host/turtle-docker /bin/turtle-docker
chmod +x /bin/turtle /bin/turtle-docker

# Add the turtle user.
useradd --no-create-home --no-user-group --shell /bin/turtle turtle

# Allow to run the turtle-docker command with sudo for the turtle user.
echo "turtle ALL = (root) NOPASSWD: /bin/turtle-docker" > /etc/sudoers.d/turtle

# Start the turtle service.
systemctl start turtle