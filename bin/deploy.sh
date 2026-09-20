#!/bin/sh
ENV=$1
shift
java -Dfile.encoding=UTF-8 -jar taskssh-1.2.1.jar deploy "$@" -i "${ENV}.yaml"