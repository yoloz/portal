#!/usr/bin/env bash
path=$(
  cd "$(dirname "$0")/.." || return
  pwd
)
nohup ${path}/bin/portal path > ${path}/log 2>&1 &