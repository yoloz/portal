#!/usr/bin/env bash
path=$(cd `dirname $0`;pwd)
nohup ${path}/portal "$@" > ${path}/log 2>&1 &