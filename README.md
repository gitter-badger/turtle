This is currently in development...
turtle requires a BTRFS partition mounted to /turtle to save all local data.

turtle is restoring apps which where running during the last daemon shutdown.
Automatically removes old backups.



Dependencies: GNU readline library and btrfs tools


docker run --privileged --name=turtle -v /var/run/docker.sock:/var/run/docker.sock -v /turtle:/turtle 1b5367bac09a
docker exec -i -t turtle turtle-client


## Encryption

openssl rand -base64 40
