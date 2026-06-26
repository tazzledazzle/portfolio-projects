pluginManagement {
    repositories {
        gradlePluginPortal()
        mavenCentral()
    }
}

plugins {
    id("org.gradle.toolchains.foojay-resolver-convention") version "1.0.0"
}

rootProject.name = "portfolio-project"

include("projgen")
include("rabbit-mq")
include("modular-jvm-build")
include("rest-api-test-demo")
include("bazel-multibuild")
include("online-bookstore")
include("ws-chat-fast")
include("workflow-api-demo")
include("otel-demo-stack")
include("platform-audit-template")
include("ai-best-practices-examples")
include("c0de-quality-and-analysis")
include("ci-cd-pipelines")
include("dev-ex")
include("dev-env")
include("onboarding-automation-cli")
include("forgex")

val pythonPluginCandidates =
    listOf(
        rootDir.resolve("gradle-python-plugin"),
        rootDir.resolve("../../may-portfolio-projects/gradle-python-plugin"),
    )

pythonPluginCandidates.firstOrNull { it.isDirectory }?.let { includeBuild(it) }
