def custom_dependencies():
    # Example: load kotlin rules, c++ toolchains, python
    load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

    rules_kotlin_version = "1.9.0"
    rules_kotlin_sha = "5766f1e599acf551aa56f49dab9ab9108269b03c557496c54acaf41f98e2b8d6"
    http_archive(
        name = "rules_kotlin",
        urls = ["https://github.com/bazelbuild/rules_kotlin/releases/download/v%s/rules_kotlin-v%s.tar.gz" % (rules_kotlin_version, rules_kotlin_version)],
        sha256 = rules_kotlin_sha,
    )

    load("@rules_kotlin//kotlin:repositories.bzl", "kotlin_repositories")

    kotlin_repositories()  # if you want the default. Otherwise see custom kotlinc distribution below

    load("@rules_kotlin//kotlin:core.bzl", "kt_register_toolchains")

    kt_register_toolchains()  # to use the default toolchain, otherwise see toolchains below
