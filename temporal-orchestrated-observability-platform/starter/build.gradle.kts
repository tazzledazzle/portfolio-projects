plugins {
    kotlin("jvm")
    application
}

dependencies {
    implementation(project(":workflows"))
    implementation("io.temporal:temporal-sdk:1.25.2")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.11.4")
}

application {
    mainClass.set("com.portfolio.temporalobs.starter.StarterMainKt")
}
