import com.pswidersk.gradle.python.VenvTask

plugins {
    kotlin("jvm") version "2.1.20"
    id("com.pswidersk.python-plugin") version "2.8.2"
    application
}

group = "com.tazzledazzle"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
    gradlePluginPortal()
}

dependencies {

    testImplementation(kotlin("test"))
}

pythonPlugin {
    pythonVersion = "3.12.8"
}

tasks.test {
    useJUnitPlatform()
}
kotlin {
    jvmToolchain(23)
}



tasks.register<VenvTask>("runPythonScript") {
    workingDir = projectDir.resolve("projgen")
    venvExec = "python"
    args = listOf("setup.py")


}