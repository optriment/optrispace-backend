#!/bin/bash -eu

cd $(dirname $(realpath $0))/..

echo -en "\033]0;🅞 Running...\a"

while sleep 2s; 
do 
    echo -en "\033]0;🅞 Running...\a"
    make run || true
    echo -en "\033]0;🅞 Restarting...\a"
done