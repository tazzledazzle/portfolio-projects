plugins {
    kotlin("jvm")
    kotlin("plugin.serialization")
    `java-library`
}

dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.3")

    val ktorVersion = "2.3.12"
    val otelVersion = "1.40.0"

    api("io.ktor:ktor-server-core:$ktorVersion")
    api("io.ktor:ktor-server-metrics-micrometer:$ktorVersion")
    api("io.micrometer:micrometer-registry-prometheus:1.13.4")
    api("io.opentelemetry:opentelemetry-api:$otelVersion")
    api("io.opentelemetry:opentelemetry-sdk:$otelVersion")
    api("io.opentelemetry:opentelemetry-sdk-extension-autoconfigure:$otelVersion")
    api("io.opentelemetry:opentelemetry-exporter-otlp:$otelVersion")
    api("io.opentelemetry.semconv:opentelemetry-semconv:1.25.0-alpha")
    api("net.logstash.logback:logstash-logback-encoder:7.4")
    api("ch.qos.logback:logback-classic:1.5.6")
}

kotlin {
    jvmToolchain(17)
}
