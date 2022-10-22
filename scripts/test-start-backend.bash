#!/bin/bash -eu

cd $(dirname $(realpath $0))/..

echo -en "\033]0;🅞 Running...\a"

while true;
do
    echo -en "\033]0;🅞 Running...\a"
    make test-start-backend || true
    echo -en "\033]0;🅞 Restarting...\a"
done
