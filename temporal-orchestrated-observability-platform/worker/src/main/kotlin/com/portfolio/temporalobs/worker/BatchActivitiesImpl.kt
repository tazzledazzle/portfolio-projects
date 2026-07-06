package com.portfolio.temporalobs.worker

import com.portfolio.temporalobs.workflows.BatchActivities
import com.portfolio.temporalobs.workflows.model.BatchEvalResult
import com.portfolio.temporalobs.workflows.model.DatasetItem
import org.slf4j.LoggerFactory

class BatchActivitiesImpl : BatchActivities {
    override fun loadDataset(
        datasetId: String,
        itemCount: Int,
    ): List<DatasetItem> {
        logger.info("loadDataset datasetId={} itemCount={}", datasetId, itemCount)
        return (1..itemCount).map { index ->
            DatasetItem(id = "$datasetId-item-$index", prompt = "Evaluate prompt $index for $datasetId")
        }
    }

    override fun scoreItem(item: DatasetItem): Double {
        logger.info("scoreItem id={}", item.id)
        val score = (item.id.hashCode() and 0xff) / 255.0
        return score
    }

    override fun aggregate(scores: List<Double>): BatchEvalResult {
        val mean = if (scores.isEmpty()) 0.0 else scores.sum() / scores.size
        logger.info("aggregate count={} mean={}", scores.size, mean)
        return BatchEvalResult(itemCount = scores.size, meanScore = mean)
    }

    companion object {
        private val logger = LoggerFactory.getLogger(BatchActivitiesImpl::class.java)
    }
}
