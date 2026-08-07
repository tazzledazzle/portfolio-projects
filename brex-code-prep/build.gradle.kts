plugins {
    alias(libs.plugins.kotlin.jvm)
    application
//    kotlin("jvm") version "2.4.10"
}

group = "org.example"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    implementation(libs.guava)
    testImplementation(kotlin("test"))
    testImplementation(libs.junit.jupiter.engine)
}

application {
    mainClass.set("org.example.AppKt")
}

kotlin {
    jvmToolchain(23)
    sourceSets.main {
        kotlin.srcDir("app/src/main/kotlin")

    }
    sourceSets.test {
        kotlin.srcDir("app/src/test/kotlin")
    }
}

tasks.test {
    useJUnitPlatform()
}
