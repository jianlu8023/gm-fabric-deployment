#!/bin/bash

user=root

userid=$(id -u $user)

if [ ! "$(id -u)" -eq $userid ]; then
  echo "change user to $user"
  exec gosu "$user" "$0" "$@"
fi

echo "starting app ..."
echo "use user $(whoami)"
echo "date $(date +"%Y-%m-%d %H:%M:%S")"
sleep 3

case "$1" in
  server)
    shift
    exec /myapp/server.bin "$@"
    ;;
  client)
    shift
    exec /myapp/client.bin "$@"
    ;;
  *)
    echo "Usage: $0 {server/client} [args...]..."
    exit 1
    ;;
esac
