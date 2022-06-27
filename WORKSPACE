workspace(name = "com_gihub_adgreetz_go_grpc_bazel_example")

bind(
    name = "go_package_prefix",
    actual = "//:go_package_prefix",
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# ================================================================
# Proto GRPC support
# https://github.com/rules-proto-grpc/rules_proto_grpc
# ================================================================

http_archive(
    name = "rules_proto_grpc",
    sha256 = "507e38c8d95c7efa4f3b1c0595a8e8f139c885cb41a76cab7e20e4e67ae87731",
    strip_prefix = "rules_proto_grpc-4.1.1",
    urls = ["https://github.com/rules-proto-grpc/rules_proto_grpc/archive/4.1.1.tar.gz"],
)

# Example: https://github.com/rules-proto-grpc/rules_proto_grpc/blob/master/example/grpc-gateway/gateway_grpc_library/WORKSPACE

load("@rules_proto_grpc//:repositories.bzl", "rules_proto_grpc_repos", "rules_proto_grpc_toolchains")

rules_proto_grpc_toolchains()

rules_proto_grpc_repos()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

load("@rules_proto_grpc//:repositories.bzl", "bazel_gazelle", "io_bazel_rules_go")  # buildifier: disable=same-origin-load

io_bazel_rules_go()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(
    version = "1.18.2",
)

bazel_gazelle()

load("@rules_proto_grpc//grpc-gateway:repositories.bzl", rules_proto_grpc_gateway_repos = "gateway_repos")

rules_proto_grpc_gateway_repos()

load("@grpc_ecosystem_grpc_gateway//:repositories.bzl", "go_repositories")

go_repositories()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load("//:go.bzl", "go_deps")

# gazelle:repository_macro go.bzl%go_deps
go_deps()

gazelle_dependencies()

# ================================================================
# Docker support requires rules_docker and custom docker rules
# ================================================================

# Download the rules_docker repository at release v0.12.1
# https://github.com/bazelbuild/rules_docker
http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "59536e6ae64359b716ba9c46c39183403b01eabfbd57578e84398b4829ca499a",
    strip_prefix = "rules_docker-0.22.0",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.22.0/rules_docker-v0.22.0.tar.gz"],
)

# This is NOT needed when going through the language lang_image
# "repositories" function(s).

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

# This requires rules_docker to be fully instantiated before
# it is pulled in.
# Download the rules_k8s repository at release v0.3.1
# https://github.com/bazelbuild/rules_k8s
http_archive(
    name = "io_bazel_rules_k8s",
    sha256 = "773aa45f2421a66c8aa651b8cecb8ea51db91799a405bd7b913d77052ac7261a",
    strip_prefix = "rules_k8s-0.5",
    urls = ["https://github.com/bazelbuild/rules_k8s/archive/v0.5.tar.gz"],
)

load("@io_bazel_rules_k8s//k8s:k8s.bzl", "k8s_defaults", "k8s_repositories")

k8s_repositories()

load("@io_bazel_rules_k8s//k8s:k8s_go_deps.bzl", k8s_go_deps = "deps")

k8s_go_deps()

k8s_defaults(
    # This becomes the name of the @repository and the rule
    # you will import in your BUILD files.
    name = "k8s_deploy",
    # This is the name of the cluster as it appears in:
    #   kubectl config current-context
    cluster = "$(cluster)",
    context = "$(cluster)",
    image_chroot = "ghcr.io/adgreetz/go-grpc-bazel-example/{ENV}",
    kind = "deployment",
    namespace = "$(namespace)",
)

# We use go_image to build a sample service
load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

_go_image_repos()

load("@io_bazel_rules_go//extras:embed_data_deps.bzl", "go_embed_data_dependencies")

go_embed_data_dependencies()

# gazelle:repo bazel_gazelle

# Golink for Gazelle
http_archive(
    name = "golink",
    sha256 = "ea728cfc9cb6e2ae024e1d5fbff185224592bbd4dad6516f3cc96d5155b69f0d",
    strip_prefix = "golink-1.0.0",
    urls = ["https://github.com/nikunjy/golink/archive/v1.0.0.tar.gz"],
)
