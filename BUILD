load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_docker//docker:docker.bzl", "docker_bundle")
load(
    "@io_bazel_rules_docker//contrib:push-all.bzl",
    docker_pushall = "docker_push",
)

package(default_visibility = ["@//visibility:public"])

# gazelle:prefix github.com/AdGreetz/go-grpc-bazel-example
gazelle(
    name = "gazelle",
    external = "external",
)

filegroup(
    name = "coverage_files",
    srcs = glob(["bazel-out/**/coverage.dat"]),
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
