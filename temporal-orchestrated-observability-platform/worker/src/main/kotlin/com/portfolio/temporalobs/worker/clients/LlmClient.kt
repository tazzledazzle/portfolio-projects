package com.portfolio.temporalobs.worker.clients

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import com.fasterxml.jackson.module.kotlin.registerKotlinModule
import com.portfolio.temporalobs.worker.telemetry.Metrics
import com.portfolio.temporalobs.workflows.model.LlmCompletion
import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.instrumentation.okhttp.v3_0.OkHttpTelemetry
import okhttp3.Call
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit

class LlmClient(
    private val baseUrl: String = System.getenv("LLM_STUB_URL") ?: "http://localhost:8090",
    private val callFactory: Call.Factory,
    private val mapper: ObjectMapper = ObjectMapper().registerKotlinModule(),
) {
    fun complete(
        question: String,
        chunks: List<String>,
    ): LlmCompletion {
        val payload =
            mapOf(
                "model" to "stub",
                "messages" to
                    listOf(
                        mapOf("role" to "system", "content" to chunks.joinToString("\n---\n")),
                        mapOf("role" to "user", "content" to question),
                    ),
            )
        val body = mapper.writeValueAsString(payload).toRequestBody(JSON)
        val request =
            Request.Builder()
                .url("${baseUrl.trimEnd('/')}/v1/chat/completions")
                .post(body)
                .build()

        val startNanos = System.nanoTime()
        var status = "ok"
        try {
            callFactory.newCall(request).execute().use { response ->
                if (!response.isSuccessful) {
                    status = "error"
                    error("LLM stub returned HTTP ${response.code}: ${response.body?.string()}")
                }
                val raw = response.body?.string() ?: error("Empty LLM stub response")
                return mapper.readValue(raw)
            }
        } catch (e: Exception) {
            status = "error"
            throw e
        } finally {
            val durationSeconds =
                TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startNanos) / 1000.0
            Metrics.recordLlmDuration(status, durationSeconds)
        }
    }

    companion object {
        private val JSON = "application/json".toMediaType()

        fun create(openTelemetry: OpenTelemetry): LlmClient =
            LlmClient(callFactory = instrumentedCallFactory(openTelemetry))

        fun instrumentedCallFactory(openTelemetry: OpenTelemetry): Call.Factory {
            val baseClient =
                OkHttpClient.Builder()
                    .connectTimeout(10, TimeUnit.SECONDS)
                    .readTimeout(30, TimeUnit.SECONDS)
                    .build()
            return OkHttpTelemetry.create(openTelemetry).newCallFactory(baseClient)
        }
    }
}
