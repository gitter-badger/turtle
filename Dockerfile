#
# DesertBit Turtle Dockerfile
#

FROM golang

MAINTAINER Roland Singer, roland.singer@desertbit.com

# Install dependencies.
RUN export DEBIAN_FRONTEND=noninteractive; \
	apt-get update && \
	apt-get install -y libreadline-dev btrfs-tools && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/desertbit/turtle

# Build and install the daemon and client.
# Finally remove the unneeded source again.
RUN go get github.com/desertbit/turtle/daemon && \
	go install github.com/desertbit/turtle/daemon && \
	go get github.com/desertbit/turtle/client && \
	go install github.com/desertbit/turtle/client && \
	rm -r /go/src

# Add additional scripts.
ADD ./scripts/turtle-crypt /usr/bin/turtle-crypt
ADD ./scripts/turtle-client /usr/bin/turtle-client
RUN chmod +x /usr/bin/turtle-crypt /usr/bin/turtle-client

# Add the ssh client user
RUN useradd --no-create-home --no-user-group --shell /go/bin/client turtle

EXPOSE 28239

VOLUME ["/turtle"]

CMD ["/go/bin/daemon"]