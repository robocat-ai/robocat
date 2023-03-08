#!/bin/bash

if [ ! -f .env ]; then
    cp .env.example .env
fi

docker build -t robocat .

docker run \
    -it \
    --rm \
    --name robocat \
    --env-file .env \
    -p ${WEB_PORT:-3000}:80 \
    -p ${VNC_PORT:-5900}:5900 \
    -v $(pwd)/flow:/home/robocat/flow \
    robocat
