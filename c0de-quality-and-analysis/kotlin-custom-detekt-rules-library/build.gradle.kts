plugins {
    kotlin("jvm") version "1.9.24"
    id("io.gitlab.arturbosch.detekt") version "1.23.6"
}

group = "com.portfolio.detekt"
version = "1.0.0"

repositories {
    mavenCentral()
}

dependencies {
    compileOnly("io.gitlab.arturbosch.detekt:detekt-api:1.23.6")
    detektPlugins(project)
    testImplementation(kotlin("test"))
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
}

detekt {
    buildUponDefaultConfig = true
    config.setFrom(files("config/detekt-custom-rules.yml"))
}
