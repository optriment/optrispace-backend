#!/bin/bash -eu

# specify script name as the first script paramameter

SCRDIR=./scripts
MBIN=./.bin/migrate

mkdir -p $SCRDIR
make $MBIN

$MBIN create -dir $SCRDIR -ext sql $1
