#!/bin/bash -eu

cd $(dirname $(realpath $0))/..
echo -en "\033]0;⏱ Watching $(pwd)...\a"
#⌚ 
while inotifywait -e close_write -r . --exclude '(\.git)|(testdata)' ; 
do 
    echo -en "\033]0;⏯ Sending stop signal!\a"
    sleep 1s
    killall -TERM work || true
    echo -en "\033]0;⏱ Watching $(pwd)...\a"
done