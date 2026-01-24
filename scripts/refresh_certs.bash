#!/usr/bin/env bash

# THIS SCRIPT ASSUMES DAILY REFRESH HAS RECENTLY SUCCEEDED
set -euf -o pipefail

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi
cd /mnt/usb/web-drones
docker compose stop nginx
certbot renew 
docker compose start nginx
