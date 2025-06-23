#!/usr/bin/env bash

set -e

echo "Copying pattern-util script from common"
cp /wd/common/scripts/pattern-util.sh pattern.sh
OLD_CONTAINER="quay.io/hybridcloudpatterns/utility-container"
NEW_CONTAINER="quay.io/dminnear/common-utility-container"
sed -i "s|$OLD_CONTAINER|$NEW_CONTAINER|" pattern.sh

if [[ "$USE_SECRETS" != "false" ]]; then
  echo "Copying template file for secrets"
  cp /wd/values-secret.yaml.template .
fi

echo "Running patternizer command to create pattern's values yaml files"
/wd/patternizer
