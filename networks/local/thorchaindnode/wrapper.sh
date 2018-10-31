#!/usr/bin/env sh

##
## Input parameters
##
BINARY=/thorchaind/${BINARY:-thorchaind}
ID=${ID:-0}
LOG=${LOG:-thorchaind.log}

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'thorchaind' E.g.: -e BINARY=thorchaind_my_test_version"
	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

##
## Run binary with all parameters
##
export THORCHAINDHOME="/thorchaind/node${ID}/thorchaind"

if [ -d "`dirname ${THORCHAINDHOME}/${LOG}`" ]; then
  "$BINARY" --home "$THORCHAINDHOME" "$@" | tee "${THORCHAINDHOME}/${LOG}"
else
  "$BINARY" --home "$THORCHAINDHOME" "$@"
fi

chmod 777 -R /thorchaind

