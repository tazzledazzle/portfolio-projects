plugins {
    kotlin("jvm") version "2.0.21"
    application
}

group = "com.patterns"
version = "1.0.0"

repositories {
    mavenCentral()
}

val resilience4jVersion = "1.7.1"

dependencies {
    // Resilience4j — application-layer circuit breaker / retry (successor to Hystrix)
    implementation("io.github.resilience4j:resilience4j-circuitbreaker:$resilience4jVersion")
    implementation("io.github.resilience4j:resilience4j-retry:$resilience4jVersion")
    implementation("io.github.resilience4j:resilience4j-decorators:$resilience4jVersion")

    // SLF4J + Logback for structured logging
    implementation("org.slf4j:slf4j-api:2.0.13")
    implementation("ch.qos.logback:logback-classic:1.5.6")

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("io.github.resilience4j:resilience4j-test:$resilience4jVersion")
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.3")
    testImplementation("org.assertj:assertj-core:3.26.0")
    testImplementation("io.mockk:mockk:1.13.11")
}

application {
    mainClass.set("com.patterns.circuitbreaker.CircuitBreakerDemoKt")
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
}
