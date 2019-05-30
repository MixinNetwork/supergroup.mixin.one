#! /usr/bin/env bash

nohup ./supergroup.mixin.one > ./httpd.log 2>&1 &
nohup ./supergroup.mixin.one -service message > ./msg.log 2>&1 & 