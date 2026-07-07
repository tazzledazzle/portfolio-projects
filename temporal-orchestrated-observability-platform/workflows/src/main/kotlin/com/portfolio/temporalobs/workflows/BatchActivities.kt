package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.BatchEvalResult
import com.portfolio.temporalobs.workflows.model.DatasetItem
import io.temporal.activity.ActivityInterface
import io.temporal.activity.ActivityMethod

@ActivityInterface
interface BatchActivities {
    @ActivityMethod
    fun loadDataset(
        datasetId: String,
        itemCount: Int,
    ): List<DatasetItem>

    @ActivityMethod
    fun scoreItem(item: DatasetItem): Double

    @ActivityMethod
    fun aggregate(scores: List<Double>): BatchEvalResult
}
