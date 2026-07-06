package com.portfolio.temporalobs.worker.clients

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import com.fasterxml.jackson.module.kotlin.registerKotlinModule

class RetrievalFixture(
    private val mapper: ObjectMapper = ObjectMapper().registerKotlinModule(),
) {
    private val chunks: List<ChunkRecord> by lazy { loadChunks() }

    fun embedQuery(question: String): String = "emb-${question.hashCode() and 0x7fffffff}"

    fun vectorSearch(embedding: String): List<String> {
        val ranked = chunks.sortedByDescending { chunk -> (embedding.hashCode() xor chunk.id.hashCode()) and 0xff }
        return ranked.take(2).map { "${it.path}#${it.id}: ${it.text}" }
    }

    private fun loadChunks(): List<ChunkRecord> {
        val stream =
            requireNotNull(javaClass.getResourceAsStream("/fixtures/chunks.json")) {
                "Missing classpath resource /fixtures/chunks.json"
            }
        return stream.use { mapper.readValue(it) }
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    data class ChunkRecord(
        val id: String,
        val path: String,
        val text: String,
    )
}
