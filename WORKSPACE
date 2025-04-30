workspace(name = "multi_lang_monorepo")

load("//bazel/custom_rules:dependencies.bzl", "custom_dependencies")

custom_dependencies()

load("//bazel/custom_rules:my_rules.bzl", "multi_lang_binary")

multi_lang_binary(
    name = "dist",
    srcs = [
        "//cpp-lib:cpp_lib",
        "//kotlin-module:kotlin_bin",
        "//python-lib:py_bin",
    ],
)
