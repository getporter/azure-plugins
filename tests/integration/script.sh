#!/bin/bash
set -euo pipefail
set +x

# This test requires users to set up an azure service principle that has access
# to an azure keyvault.
# After setting AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET environment
# variables, users also needs to set up a environment variable
# PORTER_TEST_VAULT that contains the name of the azure keyvault. Then they can run
# script like so:
# ./tests/integration/script.sh
# This script assumes users are running it from the root directory of the azure-plugin
# repo

TMP=$(mktemp -d -t tmp.XXXXXXXXXX)
PORTER_HOME=${TMP}/bin/

cleanup(){
    ret=$?
    echo "EXIT STATUS: $ret"
    rm -rf "$TMP"
    echo "cleaned up test successfully"
    exit "$ret"
}
trap cleanup EXIT

if ! command -v jq 2>&1 /dev/null; then
	echo "jq is required."
	exit 1
fi

authSetup=0
if [ -z ${AZURE_TENANT_ID} ]; then
    echo "AZURE_TENANT_ID is required for authentication."
	authSetup=1
fi

if [ -z ${AZURE_CLIENT_ID} ]; then
    echo "AZURE_CLIENT_ID is required for authentication."
	authSetup=1
fi

if [ -z ${AZURE_CLIENT_SECRET} ]; then
    echo "AZURE_CLIENT_SECRET is required for authentication."
	authSetup=1
fi

if [ $authSetup -eq 1 ]; then
	exit 1
fi

if [ -z $PORTER_TEST_VAULT ]; then
    echo "PORTER_TEST_VAULT is required for running this test."
	exit 1
fi

export PORTER_HOME=$PORTER_HOME
export AZURE_TENANT_ID=$AZURE_TENANT_ID
export AZURE_CLIENT_ID=$AZURE_CLIENT_ID
export AZURE_CLIENT_SECRET=$AZURE_CLIENT_SECRET

mkdir -p $PORTER_HOME/plugins/azure
cp ./bin/plugins/azure/azure $PORTER_HOME/plugins/azure/azure

curl -L https://cdn.porter.sh/latest/install-linux.sh | PORTER_HOME=$PORTER_HOME bash
PORTER_CMD="${TMP}/bin/porter --verbosity=debug"
secret_value=super-secret

cp ./tests/integration/testdata/config-test.yaml ${PORTER_HOME}/config.yaml

${PORTER_CMD} plugins list
cd ./tests/integration/testdata && ${PORTER_CMD} install --force --param password=$secret_value

id=$(${PORTER_CMD} installation runs list azure-plugin-test -o json | jq -r '.[].id' | head -1)

if [ -z ${id} ]; then
	echo "failed to get run id"
	exit 1
fi

value=$(az keyvault secret show --vault-name $PORTER_TEST_VAULT --name $id-password | jq -r '.value')

${PORTER_CMD} upgrade

if [[ $value == $secret_value ]]
then
	echo "test run successfully"
	exit 0
else
	echo "test failed"
	echo "expected to retrieve value: $secret_value from azure keyvault, but got: $value"
	exit 1
fi
