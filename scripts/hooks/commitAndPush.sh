#!/bin/bash

event="$1"

path="$2"
rdonly="$3"
title="$4"
syntax="$5"
tags="$6"

if [[ "rdonly" == "true" ]]; then
    exit
fi

cd $(dirname $path)

git add -A
git commit -m "Edits/adds sheet $title"
git push
