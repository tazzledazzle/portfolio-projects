plugins {
    kotlin("jvm")
    application
}

val openTelemetryVersion = "1.43.0"
val openTelemetryInstrumentationVersion = "2.16.0-alpha"

dependencies {
    implementation(project(":workflows"))
    implementation("io.temporal:temporal-sdk:1.25.2")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.fasterxml.jackson.core:jackson-databind:2.18.2")
    implementation("com.fasterxml.jackson.module:jackson-module-kotlin:2.18.2")
    implementation("org.slf4j:slf4j-api:2.0.16")
    runtimeOnly("ch.qos.logback:logback-classic:1.5.12")
    runtimeOnly("net.logstash.logback:logstash-logback-encoder:8.0")

    implementation(platform("io.opentelemetry:opentelemetry-bom:$openTelemetryVersion"))
    implementation("io.opentelemetry:opentelemetry-api")
    implementation("io.opentelemetry:opentelemetry-sdk")
    implementation("io.opentelemetry:opentelemetry-exporter-otlp")
    implementation("io.opentelemetry:opentelemetry-exporter-prometheus:$openTelemetryVersion-alpha")
    implementation("io.opentelemetry.instrumentation:opentelemetry-okhttp-3.0:$openTelemetryInstrumentationVersion")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.11.4")
    testImplementation("io.temporal:temporal-testing:1.25.2")
    testImplementation(platform("io.opentelemetry:opentelemetry-bom:$openTelemetryVersion"))
    testImplementation("io.opentelemetry:opentelemetry-sdk-testing")
    testImplementation("com.squareup.okhttp3:mockwebserver:4.12.0")
}

application {
    mainClass.set("com.portfolio.temporalobs.worker.WorkerMainKt")
}
