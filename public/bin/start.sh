#!/usr/bin/env bash
path=$(
  cd "$(dirname "$0")/.." || return
  pwd
)
nohup ${path}/bin/portal "${path}/config" > ${path}/log 2>&1 &