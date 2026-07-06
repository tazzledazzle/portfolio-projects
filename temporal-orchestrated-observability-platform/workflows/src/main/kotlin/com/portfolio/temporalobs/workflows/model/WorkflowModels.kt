package com.portfolio.temporalobs.workflows.model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import java.io.Serializable

@JsonIgnoreProperties(ignoreUnknown = true)
data class RagQaResult(
    val answer: String,
    val citation: String,
) : Serializable

@JsonIgnoreProperties(ignoreUnknown = true)
data class AgentToolsResult(
    val summary: String,
    val toolCalls: Int,
) : Serializable

@JsonIgnoreProperties(ignoreUnknown = true)
data class BatchEvalResult(
    val itemCount: Int,
    val meanScore: Double,
) : Serializable

@JsonIgnoreProperties(ignoreUnknown = true)
data class AgentPlan(
    val toolName: String,
    val arguments: String,
) : Serializable

@JsonIgnoreProperties(ignoreUnknown = true)
data class DatasetItem(
    val id: String,
    val prompt: String,
) : Serializable

@JsonIgnoreProperties(ignoreUnknown = true)
data class LlmCompletion(
    val text: String,
    val citation: String,
)
