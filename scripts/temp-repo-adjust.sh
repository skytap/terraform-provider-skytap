#!/usr/bin/env bash

echo "==> Temporarily adjusting repo directories ..."

export ORIG_TRAVIS_BUILD_DIR=${TRAVIS_BUILD_DIR}
mkdir ${TRAVIS_HOME}/gopath/src/github.com/skytap
cd ${TRAVIS_HOME}/gopath/src/github.com/skytap
mv ${ORIG_TRAVIS_BUILD_DIR} ${TRAVIS_HOME}/gopath/src/github.com/terraform-providers/terraform-provider-skytap
export TRAVIS_BUILD_DIR=${TRAVIS_HOME}/gopath/src/github.com/terraform-providers/terraform-provider-skytap
cd ${TRAVIS_HOME}/gopath/src/github.com/terraform-providers/terraform-provider-skytap

exit 0
