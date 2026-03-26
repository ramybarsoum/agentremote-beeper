#!/bin/sh
set -eu

cd "$(dirname "$0")"
BINARY_NAME=ai ./tools/maubuild "$@"
