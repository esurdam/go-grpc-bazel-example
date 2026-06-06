load("@gazelle//:def.bzl", "gazelle")

package(default_visibility = ["@//visibility:public"])

# gazelle:prefix github.com/esurdam/go-grpc-bazel-example

# google/api/annotations.proto (grpc-gateway's HTTP annotations) is vended by the
# @googleapis module, and its Go bindings are the published genproto package.
# Map both planes explicitly so Gazelle resolves the import to the real external
# targets instead of synthesizing bogus local //google/api labels. pb/helloworld
# is excluded below, so these apply to any standard, Gazelle-managed proto
# package added later.
# gazelle:resolve proto google/api/annotations.proto @googleapis//google/api:annotations_proto
# gazelle:resolve proto go google/api/annotations.proto @org_golang_google_genproto_googleapis_api//annotations

# pb/helloworld is hand-maintained: two Go proto variants (lean + gateway) share
# one importpath, plus custom gateway_* rules -- a layout Gazelle can't model.
# Exclude it so `gazelle update` never regenerates those rules, repoints the
# googleapis deps, or invents a competing go_library. Each consumer routes the
# shared importpath to its chosen variant with a `# gazelle:resolve` directive
# and pins the dep with a trailing `# keep`.
# gazelle:exclude pb/helloworld
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
