package(default_visibility = ["@//visibility:public"])

load("@bazel_gazelle//:def.bzl", "gazelle")
load("@bazel_gazelle//:def.bzl", "DEFAULT_LANGUAGES", "gazelle_binary")

# gazelle:resolve go github.com/grpc-ecosystem/grpc-gateway/v2/runtime @grpc_ecosystem_grpc_gateway//runtime:go_default_library

gazelle_binary(
    name = "gazelle_binary",
    languages = DEFAULT_LANGUAGES + ["@golink//gazelle/go_link:go_default_library"],
    visibility = ["//visibility:public"],
)

# gazelle:prefix github.com/AdGreetz/go-grpc-bazel-example
gazelle(
    name = "gazelle",
    external = "external",
    gazelle = "//:gazelle_binary",
)
