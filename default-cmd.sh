#!/usr/bin/env bash

set -e

if [[ "$USE_SECRETS" != "false" ]]; then
  echo "Copying template file for secrets"
  cp /wd/values-secret.yaml.template .
  sed -i 's|\${USE_SECRETS:=false}|\${USE_SECRETS:=true}|' /wd/pattern.sh
fi

cp /wd/pattern.sh .

echo "Running patternizer command to create pattern's values yaml files"
/wd/patternizer
