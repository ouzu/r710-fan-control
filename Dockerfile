FROM debian

RUN apt update && \
apt upgrade -y && \
apt install -y lm-sensors golang git && \
rm -rf /var/lib/apt/lists/*