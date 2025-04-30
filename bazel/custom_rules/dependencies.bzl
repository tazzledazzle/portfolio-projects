def custom_dependencies():
    # Example: load kotlin rules, c++ toolchains, python
    load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

    http_archive(
        name = "io_bazel_rules_kotlin",
        url = "https://github.com/bazelbuild/rules_kotlin/releases/download/v1.5.0/rules_kotlin_release.tar.gz",
        strip_prefix = "rules_kotlin-1.5.0",
    )

    load("@io_bazel_rules_kotlin//kotlin:dependencies.bzl", "kotlin_dependencies")

    kotlin_dependencies()
