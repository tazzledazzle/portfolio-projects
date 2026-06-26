import com.tazzledazzle.python.tasks.PythonExec

plugins {
    id("com.tazzledazzle.python") version "0.2.0"
    base
}


tasks.register<PythonExec>("buildEnvironmentDriftDetector") {
    executable.set("pip")
    arguments.set(listOf("install", "-e", "environment-drift-detector"))
}


tasks.register<PythonExec>("buildRemoteDevEnvironmentOrchestrator") {
    executable.set("pip")
    arguments.set(listOf("install","-e","remote-dev-environment-orchestrator"))
}
