plugins {
    kotlin("jvm") version "2.3.0"
    kotlin("multiplatform") version "2.3.0" apply false
    id("com.tazzledazzle.python") version "0.2.0" apply false

}

group = "com.tazzledazzle"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {

    testImplementation(kotlin("test"))
}

tasks.test {
    useJUnitPlatform()
}
kotlin {
    jvmToolchain(23)
}