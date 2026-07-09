plugins {
    kotlin("jvm") version "1.9.24"
    application
}

group = "com.company.onboarding"
version = "0.1.0"

repositories {
    mavenCentral()
}

dependencies {
    implementation(kotlin("stdlib"))
    testImplementation(kotlin("test"))
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
}

application {
    mainClass.set("com.company.onboarding.MainKt")
}
