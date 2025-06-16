#!/usr/bin/env bash

set -e

/wd/patternizer

cp -a /wd/common .
rm -rf ./common/.git

cp /wd/Makefile .
cp /wd/values-secret.yaml.template .

ln -s common/scripts/pattern-util.sh pattern.sh
