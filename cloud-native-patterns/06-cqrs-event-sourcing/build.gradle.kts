plugins {
    kotlin("jvm") version "2.0.21"
    application
}

group = "com.patterns"
version = "1.0.0"

repositories {
    mavenCentral()
}

dependencies {
    // No framework dependencies — this is intentionally hand-rolled to show the mechanics.
    // In production you'd add EventStoreDB client or Axon Framework here.

    // Logging
    implementation("org.slf4j:slf4j-api:2.0.13")
    implementation("ch.qos.logback:logback-classic:1.5.6")

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.3")
    testImplementation("org.assertj:assertj-core:3.26.0")
}

application {
    mainClass.set("com.patterns.cqrs.MainKt")
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
}
