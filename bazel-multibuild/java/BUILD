load("@rules_java//java:defs.bzl", "java_binary", "java_test", "java_library")

package(default_visibility = ["//visibility:public"])

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
    name = "tests",
    srcs = glob(["src/test/java/com/example/myproject/*.java"]),
    deps = [
        ":java-main-lib",
        "@maven//:com_google_guava_guava",
        "@maven//:junit_junit",
    ],
)

java_binary(
    name = "java-maven",
    main_class = "com.example.myproject.App",
) 