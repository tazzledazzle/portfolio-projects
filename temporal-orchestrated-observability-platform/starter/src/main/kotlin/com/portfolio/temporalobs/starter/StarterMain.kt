package com.portfolio.temporalobs.starter

import com.portfolio.temporalobs.workflows.AgentToolsWorkflow
import com.portfolio.temporalobs.workflows.BatchEvalWorkflow
import com.portfolio.temporalobs.workflows.PingWorkflow
import com.portfolio.temporalobs.workflows.RagQaWorkflow
import com.portfolio.temporalobs.workflows.TaskQueues
import com.portfolio.temporalobs.workflows.TemporalConnection
import com.portfolio.temporalobs.workflows.model.AgentToolsResult
import com.portfolio.temporalobs.workflows.model.BatchEvalResult
import com.portfolio.temporalobs.workflows.model.RagQaResult
import io.temporal.client.WorkflowOptions
import java.util.UUID
import kotlin.system.exitProcess

fun main(args: Array<String>) {
    when (args.firstOrNull()) {
        "ping" -> runPing()
        "rag" -> runRag(args.drop(1))
        "agent" -> runAgent(args.drop(1))
        "batch" -> runBatch(args.drop(1))
        null, "" -> {
            printUsage()
            exitProcess(1)
        }
        else -> {
            System.err.println("Unknown command: ${args.first()}")
            printUsage()
            exitProcess(1)
        }
    }
}

private fun runPing() {
    val workflowId = "ping-${UUID.randomUUID()}"
    val workflow =
        workflowStub(PingWorkflow::class.java, workflowId)
    val result = workflow.ping()
    printResult(workflowId, "result=$result")
}

private fun runRag(args: List<String>) {
    val question = args.joinToString(" ").ifBlank { "What is Temporal orchestration?" }
    val workflowId = "rag-${UUID.randomUUID()}"
    val workflow = workflowStub(RagQaWorkflow::class.java, workflowId)
    val result: RagQaResult = workflow.ask(question)
    printResult(
        workflowId,
        "answer=${result.answer}",
        "citation=${result.citation}",
    )
}

private fun runAgent(args: List<String>) {
    val goal = args.joinToString(" ").ifBlank { "Summarize observability architecture" }
    val workflowId = "agent-${UUID.randomUUID()}"
    val workflow = workflowStub(AgentToolsWorkflow::class.java, workflowId)
    val result: AgentToolsResult = workflow.run(goal)
    printResult(
        workflowId,
        "summary=${result.summary}",
        "tool_calls=${result.toolCalls}",
    )
}

private fun runBatch(args: List<String>) {
    val datasetId = args.firstOrNull() ?: "demo-eval"
    val itemCount = args.getOrNull(1)?.toIntOrNull() ?: 5
    val workflowId = "batch-${UUID.randomUUID()}"
    val workflow = workflowStub(BatchEvalWorkflow::class.java, workflowId)
    val result: BatchEvalResult = workflow.eval(datasetId, itemCount)
    printResult(
        workflowId,
        "item_count=${result.itemCount}",
        "mean_score=${result.meanScore}",
    )
}

private fun <T> workflowStub(
    workflowClass: Class<T>,
    workflowId: String,
): T {
    val client = TemporalConnection.workflowClient()
    return client.newWorkflowStub(
        workflowClass,
        WorkflowOptions.newBuilder()
            .setWorkflowId(workflowId)
            .setTaskQueue(TaskQueues.AI_WORKFLOWS)
            .build(),
    )
}

private fun printResult(
    workflowId: String,
    vararg lines: String,
) {
    println("workflow_id=$workflowId")
    lines.forEach { println(it) }
}

private fun printUsage() {
    println("Usage: starter <command> [args]")
    println("Commands:")
    println("  ping                 Run PingWorkflow")
    println("  rag [question]       Run RagQaWorkflow")
    println("  agent [goal]         Run AgentToolsWorkflow")
    println("  batch [dataset] [n]  Run BatchEvalWorkflow (default n=5)")
}
