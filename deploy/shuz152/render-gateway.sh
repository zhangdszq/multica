#!/bin/sh
set -eu
root=/opt/multica-shuz152
backend_ip=$(docker inspect multica-shuz152-backend --format '{{(index .NetworkSettings.Networks "multica-shuz152-internal").IPAddress}}')
frontend_ip=$(docker inspect multica-shuz152-frontend --format '{{(index .NetworkSettings.Networks "multica-shuz152-internal").IPAddress}}')
test -n "$backend_ip"
test -n "$frontend_ip"
sed -e "s/__BACKEND_IP__/$backend_ip/g" -e "s/__FRONTEND_IP__/$frontend_ip/g" "$root/nginx.conf.template" > "$root/nginx.conf"
/usr/sbin/nginx -t -p "$root/" -c "$root/nginx.conf"
