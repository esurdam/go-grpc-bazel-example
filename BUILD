package(default_visibility = ["@//visibility:public"])

load("@bazel_gazelle//:def.bzl", "gazelle")
load("@bazel_gazelle//:def.bzl", "DEFAULT_LANGUAGES", "gazelle_binary")

# gazelle:resolve go github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld //pb/helloworld:helloworld_gateway_lib_proto
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

filegroup(
    name = "coverage_files",
    srcs = glob(["bazel-out/**/coverage.dat"]),
)

load("@io_bazel_rules_docker//docker:docker.bzl", "docker_bundle")
load(
    "@io_bazel_rules_docker//contrib:push-all.bzl",
    docker_pushall = "docker_push",
)

docker_bundle(
    name = "bundle",
    images = {
        "ghcr.io/adgreetz/go-grpc-bazel-example:{BUILD_USER}": "//services/helloworld:docker",
    },
)

docker_pushall(
    name = "push",
    bundle = ":bundle",
)
