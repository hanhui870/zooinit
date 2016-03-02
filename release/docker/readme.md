# Golang Project Fast Build Guide

1. Build golang-docker-compiler image: docker build -t haimi:go-docker-dev .
1. Build project written in golang in dir project.
1. Pack the distribution into a new docker image for production run.
1. The data transfer between dev machine and docker container use docker volume.

