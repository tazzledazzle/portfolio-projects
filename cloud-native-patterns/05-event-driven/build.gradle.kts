plugins {
    kotlin("jvm") version "2.0.21"
    application
}

group = "com.patterns"
version = "1.0.0"

repositories {
    mavenCentral()
}

val coroutinesVersion = "1.8.1"

dependencies {
    // Coroutines for the outbox relay background job
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:$coroutinesVersion")

    // JSON serialization (simulating what you'd use for event payloads)
    implementation("com.fasterxml.jackson.module:jackson-module-kotlin:2.17.1")

    // Logging
    implementation("org.slf4j:slf4j-api:2.0.13")
    implementation("ch.qos.logback:logback-classic:1.5.6")

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.3")
    testImplementation("org.assertj:assertj-core:3.26.0")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:$coroutinesVersion")
}

application {
    mainClass.set("com.patterns.eventdriven.MainKt")
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
}
