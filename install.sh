#!/bin/bash

set -e

# Pull the turtle docker image.
docker pull desertbit/turtle

# Install the files.
cp -f ./scripts/turtle.service /etc/systemd/system/turtle.service
cp -f ./scripts/turtle /bin/turtle
chmod +x /bin/turtle

# Add the turtle user.
useradd --no-create-home --no-user-group --shell /bin/turtle turtle

# Enable and start the turtle service.
systemctl enable turtle
systemctl start turtle