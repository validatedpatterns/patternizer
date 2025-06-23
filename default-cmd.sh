#!/usr/bin/env bash

set -e

cp /wd/pattern.sh .

if [[ "$USE_SECRETS" != "false" ]]; then
  echo "Copying template file for secrets"
  cp /wd/values-secret.yaml.template .
fi

echo "Running patternizer command to create pattern's values yaml files"
/wd/patternizer
