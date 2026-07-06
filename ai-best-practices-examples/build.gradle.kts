import com.tazzledazzle.python.tasks.PythonExec

plugins {
    id("com.tazzledazzle.python") version "0.1.0"
}

repositories {
    gradlePluginPortal()
    mavenCentral()
}

python {
    pythonVersion.set("3.13.6")
}
//todo: pip install -e command

tasks.register<PythonExec>("buildAICodeAssistant").configure {
    // check if venv active
    // install deps
    // run entrypoint
}
tasks.register<PythonExec>("buildAIImageVideoGenerator").configure {
    // check if venv active
    // install deps
    // run entrypoint
}
tasks.register<PythonExec>("buildChatUI").configure {
    // check if venv active
    executable.set("${project.projectDir.parent}/.venv/bin/pip")
    arguments = listOf("install","-e","\"${project.projectDir.parent}/ai-best-practice-examples/chat-ai/\"")
    // install deps
    // run entrypoint
}
tasks.register<PythonExec>("buildDomainExpertAI").configure {
    // check if venv active
    // install deps
    // run entrypoint
}
tasks.register<PythonExec>("buildKnowledgeQASystem").configure {
    // check if venv active
    // install deps
    // run entrypoint
}
