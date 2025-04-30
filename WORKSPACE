workspace(name = "multi_lang_monorepo")

load("//bazel/custom_rules:dependencies.bzl", "custom_dependencies")

custom_dependencies()

load("//bazel/custom_rules:my_rules.bzl", "multi_lang_binary")

multi_lang_binary(
    name = "dist",
    srcs = [
        "//src/main/java/com/example:Main.java",
        "//src/main/python:main.py",
    ],
    deps = [
        "//src/main/java/com/example:lib",
        "//src/main/python:lib",
    ],
)
