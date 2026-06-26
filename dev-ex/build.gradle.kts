import com.github.gradle.node.npm.task.NpmTask
import com.tazzledazzle.python.tasks.PythonExec

plugins {
    id("com.tazzledazzle.python")
    id("com.github.node-gradle.node") version "7.1.0"
    kotlin("multiplatform")
}
repositories {
    mavenCentral()
}

python {
    pythonVersion.set("3.13.6")
}

kotlin {
    js(IR) {
        browser {
            binaries.executable()
        }
    }
}

node {
    nodeProjectDir.set(file("${project.projectDir}/tooling-adoption-tracker"))
    version.set("20.11.0")
    download.set(true)
    npmCommand = listOf("run", "build").joinToString(" ")
}
tasks.register<PythonExec>("buildDeveloperSatisfactionPulseSystem") {

    description = "dev-ex/developer-satisfaction-pulse-system build"
    arguments = listOf("-m", "pip", "install", "-e", "developer-satisfaction-pulse-system/")
    executable.set("python")
}
tasks.register<PythonExec>("buildInnerLoopFrictionScorer") {

    description = "dev-ex/inner-loop-friction-scorer build"
    arguments = listOf("-m", "pip", "install", "-e", "inner-loop-friction-scorer/")
    executable.set("python")
}
tasks.register<PythonExec>("buildPlatformChangelogMigrationGenerator") {

    description = "dev-ex/platform-changelog-migration-generator build"
    arguments = listOf("-m", "pip", "install", "-e", "platform-changelog-migration-generator/")
    executable.set("python")
}

tasks.named("build") {
    dependsOn(
        "buildPlatformChangelogMigrationGenerator",
        "jsBrowserDevelopmentExecutableDistribution",
        "buildInnerLoopFrictionScorer",
        "buildDeveloperSatisfactionPulseSystem",
    )
}

val buildTaskUsingNpm =
    tasks.register<NpmTask>("buildNpm") {
        dependsOn(tasks.npmInstall)
        npmCommand.set(listOf("run", "build"))
        args.set(listOf("--", "--outDir", "${project.projectDir}/tooling-adoption-tracker/npm-output"))
        inputs.dir("tooling-adoption-tracker/src")
        outputs.dir("${project.projectDir}/tooling-adoption-tracker/npm-output")
    }
