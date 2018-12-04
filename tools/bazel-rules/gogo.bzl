load("@io_bazel_rules_go//go:def.bzl", "go_repository")


_BUILD_FILE = """
proto_library(
    name = "gogoproto",
    srcs = ["gogo.proto"],
    deps = [
        "@com_google_protobuf//:descriptor_proto",
    ],
    visibility = ["//visibility:public"],
)
"""

def _go_repository_impl(ctx):
  ctx.file("BUILD.bazel", content="")
  ctx.file("github.com/gogo/protobuf/gogoproto/BUILD.bazel", content=_BUILD_FILE)
  ctx.template("github.com/gogo/protobuf/gogoproto/gogo.proto", ctx.attr._proto)

_gogo_proto_repository = repository_rule(
    implementation = _go_repository_impl,
    attrs = {
        "_proto": attr.label(default="//vendor/github.com/gogo/protobuf/gogoproto:gogo.proto"),
    },
)

def gogo_dependencies():
  go_repository(
      name = "com_github_gogo_protobuf",
      importpath = "github.com/gogo/protobuf",
      urls = ["https://codeload.github.com/ianthehat/protobuf/zip/2adc21fd136931e0388e278825291678e1d98309"],
      strip_prefix = "protobuf-2adc21fd136931e0388e278825291678e1d98309",
      type = "zip",
      build_file_proto_mode="disable",
  )
  _gogo_proto_repository(name = "internal_gogo_proto_repository")
