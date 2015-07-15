# Turtle - Rock Solid Cluster Management

[![Join the chat at https://gitter.im/desertbit/turtle](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/desertbit/turtle?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

***This application is under development. It is stable to use. Detailed documentation will follow***

Turtle is a docker and cluster management tool. It manages Turtle Apps which can be deployed within seconds. It offers an unique and powerful backup system, based on BTRFS snapshots.

## Requirements

Turtle requires a BTRFS partition mounted to /turtle.
All apps and backups are stored in that location.

# Installation


```
docker pull desertbit/turtle
```

docker exec -i -t turtle turtle-client


