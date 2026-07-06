package com.portfolio.temporalobs.worker

import com.portfolio.temporalobs.workflows.RagActivities
import com.portfolio.temporalobs.worker.clients.LlmClient
import com.portfolio.temporalobs.worker.clients.RetrievalFixture
import org.slf4j.LoggerFactory

class RagActivitiesImpl(
    private val retrieval: RetrievalFixture = RetrievalFixture(),
    private val llmClient: LlmClient,
) : RagActivities {
    override fun embedQuery(question: String): String {
        logger.info("embedQuery questionLength={}", question.length)
        return retrieval.embedQuery(question)
    }

    override fun vectorSearch(embedding: String): String {
        logger.info("vectorSearch embeddingPrefix={}", embedding.take(12))
        return retrieval.vectorSearch(embedding).joinToString("\n---\n")
    }

    override fun llmComplete(
        question: String,
        chunks: String,
    ): String {
        val chunkList = chunks.split("\n---\n").filter { it.isNotBlank() }
        logger.info("llmComplete chunks={}", chunkList.size)
        val completion = llmClient.complete(question, chunkList)
        return "${completion.text}||${completion.citation}"
    }

    companion object {
        private val logger = LoggerFactory.getLogger(RagActivitiesImpl::class.java)
    }
}
