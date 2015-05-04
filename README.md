# Turtle - Rock Solid Cluster Management

Turtle is a docker and cluster management tool. It manages Turtle Apps which can be deployed within seconds. It offers an unique and powerful backup system, based on BTRFS snapshots.

## Requirements

Turtle requires a BTRFS partition mounted to /turtle.
All apps and backups are stored in that location.

# Installation


```
docker pull desertbit/turtle
```

docker exec -i -t turtle turtle-client


