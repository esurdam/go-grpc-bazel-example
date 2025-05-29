load("@gazelle//:def.bzl", "gazelle")

package(default_visibility = ["@//visibility:public"])

# gazelle:prefix github.com/esurdam/go-grpc-bazel-example
gazelle(
    name = "gazelle",
    external = "external",
)

#filegroup(
#    name = "coverage_files",
#    srcs = glob(["bazel-out/**/coverage.dat"]),
#)

filegroup(
    name = "build_all",
    srcs = [
        "//cmd/helloworld-client",
        "//services/helloworld",
    ],
)

filegroup(
    name = "push_all",
    srcs = [
        "//cmd/helloworld-client:push",
        "//services/helloworld:push",
    ],
)
