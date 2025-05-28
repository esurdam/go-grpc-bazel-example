//go:build !bazel
// +build !bazel

package main

import _ "embed"

//go:embed helloworld_openapi_swagger.json
var Data []byte
