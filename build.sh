#!/bin/sh
set -e

version=$(git describe --tags --always --dirty 2>/dev/null || printf 'dev')
exec go build -ldflags "-s -w -X github.com/MixinNetwork/supergroup.mixin.one/config.BuildVersion=$version" "$@"
