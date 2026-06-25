import com.tazzledazzle.python.tasks.PythonExec

plugins {
    id("com.tazzledazzle.python") version "0.2.0"
}


tasks.register<PythonExec>("buildApiBreakingChangeDetector").configure {
    description = "api-breaking-change-detector build"
}