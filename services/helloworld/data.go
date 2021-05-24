// +build local

package main

// Data is used to represent the bazel autocompiled Data for `swagger.json`. Used to satisfy
// Data dep when NOT using bazel.
var Data []byte
