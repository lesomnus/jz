export PATH="$PATH:$(go env GOROOT)/lib/wasm"
export GOOS="js"
export GOARCH="wasm"

export JZ_TEST_FETCH_TARGET_URL="http://localhost:7743"
