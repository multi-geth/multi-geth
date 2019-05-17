#!/usr/bin/env bash

set -e

# Check architecture: currently we only support EVM-RS enabled builds on x86_64 and i684
MACHINE_TYPE=`uname -m`
case ${MACHINE_TYPE} in
    x86_64|i686)
        echo "Architecture $MACHINE_TYPE is supported, continuing"
        ;;
    *)
        echo "EVM-RS enabled builds are only supported on x86_64 and i686 at the moment."
        exit 1
        ;;
esac

installrust() {
    curl https://sh.rustup.rs -sSf | sh
    source $HOME/.cargo/env
}

# Prompt to install rust and cargo if not already installed.
if hash cargo 2>/dev/null; then
    echo "Cargo installed OK, continuing"
else
    while true; do
        read -p "Install/build with EVM-RS requires Rust and cargo to be installed. Would you like to install them? [Yy|Nn]" yn
        case $yn in
            [yY]* ) installrust; echo "Rust and cargo have been installed and temporarily added to your PATH"; break;;
            [nN]* ) echo "Can't compile EVM-RS. Exiting."; exit 0;;
        esac
    done
fi

OS=`uname -s`

geth_path="github.com/ethereum/go-ethereum"
sputnik_path="github.com/ethereumproject/evm-ffi"
sputnik_dir="$GOPATH/src/$geth_path/vendor/$sputnik_path"

echo "Building EVM-RS"
make -C "$sputnik_dir/c"

LDFLAGS="$sputnik_dir/c/libsputnikvm.a "
case $OS in
	"Linux")
		LDFLAGS+="-ldl -lm"
		;;

	"Darwin")
		LDFLAGS+="-ldl -lresolv -lm"
		;;

    CYGWIN*|MINGW32*|MSYS*)
		LDFLAGS="-Wl,--allow-multiple-definition $sputnik_dir/c/sputnikvm.lib -lws2_32 -luserenv -lm"
		;;
esac

export CGO_LDFLAGS=$LDFLAGS