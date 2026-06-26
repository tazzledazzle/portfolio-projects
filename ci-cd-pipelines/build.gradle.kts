import com.tazzledazzle.python.tasks.PythonExec

plugins {
    id("com.tazzledazzle.python")
    base
}
tasks.register<Exec>("activateVenv").configure {
    commandLine = listOf("../.venv/bin/activate")
}

tasks.register<PythonExec>("buildPipelineCostAnalyzer").configure {
    executable.set("python")
    listOf("src/main.py")
}
// build.gradle.kts


// ── Configuration ────────────────────────────────────────────────────────────

val venvDir = layout.projectDirectory.dir(".venv")
val isWindows = System.getProperty("os.name").lowercase().contains("win")

val pythonExecutable: String
    get() = if (isWindows)
        venvDir.file("Scripts/python.exe").asFile.absolutePath
    else
        venvDir.file("bin/python").asFile.absolutePath

val pipExecutable: String
    get() = if (isWindows)
        venvDir.file("Scripts/pip.exe").asFile.absolutePath
    else
        venvDir.file("bin/pip").asFile.absolutePath

// ── Helper: build a command that runs inside the venv ────────────────────────

/**
 * Returns a cross-platform command list that activates the venv
 * and then runs [args] inside it.
 */
fun venvCommand(vararg args: String): List<String> =
    if (isWindows)
        listOf("cmd", "/c", venvDir.file("Scripts/activate.bat").asFile.absolutePath, "&&") + args.toList()
    else
        listOf("bash", "-c", "source ${venvDir.file("bin/activate").asFile.absolutePath} && ${args.joinToString(" ")}")

// ── Tasks ─────────────────────────────────────────────────────────────────────

/**
 * Creates the virtual environment if it doesn't already exist.
 */
val createVenv by tasks.registering(Exec::class) {
    group = "python"
    description = "Create the Python virtual environment in .venv/"

    onlyIf { !venvDir.asFile.exists() }

    commandLine("python3", "-m", "venv", venvDir.asFile.absolutePath)
}

/**
 * Installs dependencies from requirements.txt inside the venv.
 */
val installDeps by tasks.registering(Exec::class) {
    group = "python"
    description = "Install Python dependencies from requirements.txt"

    dependsOn(createVenv)

    val requirementsFile = layout.projectDirectory.file("requirements.txt").asFile
    onlyIf { requirementsFile.exists() }

    commandLine(pipExecutable, "install", "-r", requirementsFile.absolutePath)
}

/**
 * Activates the venv and runs your main Python script.
 * Swap out the script name / args to suit your project.
 */
val runCanaryDeploymentController by tasks.registering(Exec::class) {
    group = "python"
    description = "Run main.py inside the activated virtual environment"

    dependsOn(installDeps)

    // Option A – call the venv's python binary directly (simpler, recommended)
    commandLine(pythonExecutable, "canary-deployment-controller/src/main.py")

    // Option B – activate the venv shell-style then run (uncomment to use instead)
    // commandLine(*venvCommand("python", "main.py").toTypedArray())
}
val runFlakyPipelineGate by tasks.registering(Exec::class) {
    group = "python"
    description = "Run main.py inside the flaky pipeline gate project"

    dependsOn(installDeps)

    commandLine(pythonExecutable, "flaky-pipeline-gate/src/main.py")
}


val runPipelineCostAnalyzer by tasks.registering(Exec::class) {
    group = "python"
    description = "Run main.py inside the pipeline cost analyzer project"

    dependsOn(installDeps)

    commandLine(pythonExecutable, "pipeline-cost-analyzer/src/main.py")
}


val runPipelineTelemetryExporter by tasks.registering(Exec::class) {
    group = "python"
    description = "Run main.py inside the pipeline cost analyzer project"

    dependsOn(installDeps)

    commandLine(pythonExecutable, "pipeline-telemetry-exporter/src/main.py")
}


val runReleaseLeadTimeCalculator by tasks.registering(Exec::class) {
    group = "python"
    description = "Run main.py inside the release lead time calculator project"

    dependsOn(installDeps)

    commandLine(pythonExecutable, "release-lead-time-calculator/src/main.py")
}


val runSelfServicePipelineTemplateEngine by tasks.registering(Exec::class) {
    group = "python"
    description = "Run main.py inside the self service pipeline template engine project"

    dependsOn(installDeps)

    commandLine(pythonExecutable, "self-service-pipeline-template-engine/src/main.py")
}
/**
 * General-purpose task: activate the venv and run any command.
 *
 * Usage:
 *   ./gradlew venvExec -PvenvArgs="pytest tests/ -v"
 */
val venvExec by tasks.registering(Exec::class) {
    group = "python"
    description = "Run an arbitrary command inside the venv. Pass -PvenvArgs='<cmd>'"

    dependsOn(createVenv)

    val rawArgs = providers.gradleProperty("venvArgs").orNull
        ?: error("Pass -PvenvArgs='<your command>' to use this task")

    commandLine(*venvCommand(rawArgs).toTypedArray())
}

/**
 * Prints the venv's Python version to confirm activation works.
 */
val verifyVenv by tasks.registering(Exec::class) {
    group = "python"
    description = "Verify the venv is active and print its Python version"

    dependsOn(createVenv)

    commandLine(pythonExecutable, "--version")
}

tasks.named("build").configure {
    dependsOn(
        listOf(
            runSelfServicePipelineTemplateEngine,
            runReleaseLeadTimeCalculator,
            runPipelineTelemetryExporter,
            runPipelineCostAnalyzer,
            runFlakyPipelineGate,
            runCanaryDeploymentController
        )
    )
}
