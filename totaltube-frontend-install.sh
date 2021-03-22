#!/usr/bin/env bash

set -e

# Downloading right totaltube-frontend binary

if [[ "$OSTYPE" == "freebsd"* ]]; then
    curl -s https://totaltraffictrader.com/latest/freebsd/totaltube-frontend.tar.gz | tar xvzf - -C /usr/local/bin
else
    curl -s https://totaltraffictrader.com/latest/linux/totaltube-frontend.tar.gz | tar xvzf - -C /usr/local/bin
fi


# Starting installation
/usr/local/bin/totaltube-frontend install < /dev/tty