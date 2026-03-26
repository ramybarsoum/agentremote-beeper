#!/bin/sh

agentremote_resolve_go_crypto_backend() {
	AGENTREMOTE_GO_TAG=""
	agentremote_crypto_backend="${AGENTREMOTE_CRYPTO_BACKEND:-goolm}"

	case "$agentremote_crypto_backend" in
		goolm)
			AGENTREMOTE_GO_TAG="goolm"
			;;
		libolm)
			;;
		*)
			printf '%s\n' "error: unsupported AGENTREMOTE_CRYPTO_BACKEND '$agentremote_crypto_backend' (expected 'goolm' or 'libolm')" >&2
			return 1
			;;
	esac
}
