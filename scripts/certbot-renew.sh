#!/bin/sh
set -eu

: "${CERTBOT_EMAIL:?Set CERTBOT_EMAIL before starting the prod profile}"

while :; do
    if certbot certonly --webroot --webroot-path /var/www/certbot \
        --cert-name webdrones.net -d webdrones.net \
        --email "$CERTBOT_EMAIL" \
        --agree-tos --non-interactive --keep-until-expiring \
        --deploy-hook 'kill -HUP 1'; then
        if [ ! -e /etc/nginx/tls/webdrones.conf ]; then
            ln -s /etc/nginx/tls-template/webdrones.conf /etc/nginx/tls/webdrones.conf
            # The initial HTTP-only nginx config can start without certificate files.
            kill -HUP 1
        fi
    fi

    # Certbot checks on startup and every 12 hours; it renews only when needed.
    sleep 12h
done
