plugins {
    kotlin("jvm") version "2.0.21"
    application
}

group = "com.patterns"
version = "1.0.0"

repositories {
    mavenCentral()
}

val temporalVersion = "1.25.2"

dependencies {
    // Temporal SDK
    implementation("io.temporal:temporal-sdk:$temporalVersion")

    // Logging
    implementation("org.slf4j:slf4j-api:2.0.13")
    implementation("ch.qos.logback:logback-classic:1.5.6")

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("io.temporal:temporal-testing:$temporalVersion")
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.3")
    testImplementation("org.assertj:assertj-core:3.26.0")
    testImplementation("io.mockk:mockk:1.13.11")
}

application {
    mainClass.set("com.patterns.saga.SagaWorkerKt")
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
}
