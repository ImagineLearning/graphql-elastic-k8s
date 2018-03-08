#!/bin/bash

sh ./scripts/build_linux.sh
docker-compose build
docker tag graphql_graphql:latest ilprovo/graphql:latest
docker push ilprovo/graphql:latest
