load("@rules_java//java:defs.bzl", "java_binary", "java_library", "java_test")

package(default_visibility = ["//visibility:public"])

java_library(
    name = "greeter",
    srcs = ["src/main/java/com/example/Greeting.java"],
)

java_binary(
    name = "ProjectRunner",
    srcs = glob(["src/main/java/com/example/*.java"]),
)

java_library(
    name = "java-main-lib",
    srcs = glob(["src/main/java/com/example/myproject/*.java"]),
    deps = [
        "@maven//:com_google_guava_guava",
    ]
)


java_test(
    name = "myproject_test",
    srcs = glob(["src/test/java/com/example/myproject/*.java"]),
    test_class = "com.example.myproject.TestApp",
    deps = [
        ":java-main-lib",
        "@maven//:com_google_guava_guava",
        "@maven//:junit_junit",
    ],
)

java_binary(
    name = "java-maven",
    main_class = "com.example.myproject.App",
    runtime_deps = [":java-main-lib"],
)

# Container image targets require rules_oci extension wiring; see MODULE.bazel.
# Uncomment when oci.pull is validated on your Bazel version:
#
# oci_image(
#     name = "image",
#     ...
# )