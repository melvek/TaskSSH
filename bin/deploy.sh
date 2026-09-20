#!/bin/sh
ENV=$1
shift
java -Dfile.encoding=UTF-8 -jar taskssh-1.2.0.jar deploy "$@" -i "${ENV}.yaml"