#!/bin/bash
#set -eo pipefail

cd ~/ws/ezpkg/_examples
output=$(./getall.sh 2>&1)
exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo ":checkmark.seal.fill:"
    echo "---"
    $SWIFTBAR_PLUGINS_PATH/../swiftbar.ezpkg/_cli "$output"
else
    echo ":multiply.circle:"
    echo "---"
    $SWIFTBAR_PLUGINS_PATH/../swiftbar.ezpkg/_cli "$output"
fi
