plugins {
    kotlin("jvm") version "1.9.22"
}

kotlin {
    jvmToolchain(17)
}

dependencies {
    implementation(project(":core"))
    implementation("org.springframework.boot:spring-boot-starter-web:3.2.2")
    compileOnly("org.springframework.boot:spring-boot-starter:3.2.2")
}

tasks.test {
    useJUnitPlatform()
}
