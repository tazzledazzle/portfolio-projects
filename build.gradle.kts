plugins {
    kotlin("jvm") version "2.3.0"
    kotlin("multiplatform") version "2.3.0" apply false
    id("org.jlleitschuh.gradle.ktlint") version "14.2.0"
    id("io.gitlab.arturbosch.detekt") version "1.23.8" apply false
    id("com.tazzledazzle.python") version "0.2.0" apply false
}

group = "com.tazzledazzle"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
    gradlePluginPortal()
}

dependencies {
    testImplementation(kotlin("test"))
}

tasks.test {
    useJUnitPlatform()
}

kotlin {
    jvmToolchain(23)
}

ktlint {
    version.set("1.5.0")
    android.set(false)
    filter {
        exclude { element -> element.file.path.contains("${File.separator}bin${File.separator}") }
    }
}

subprojects {
    pluginManager.withPlugin("org.jetbrains.kotlin.jvm") {
        apply(plugin = "org.jlleitschuh.gradle.ktlint")
    }
    pluginManager.withPlugin("org.jetbrains.kotlin.multiplatform") {
        apply(plugin = "org.jlleitschuh.gradle.ktlint")
    }
}
