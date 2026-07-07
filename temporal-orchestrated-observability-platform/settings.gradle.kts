pluginManagement {
    repositories {
        mavenCentral()
        gradlePluginPortal()
    }
}

rootProject.name = "temp-orch-obse-plat"

include("workflows")
include("worker")
include("starter")
