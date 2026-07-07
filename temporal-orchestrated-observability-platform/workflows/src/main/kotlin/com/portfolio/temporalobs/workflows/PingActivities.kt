package com.portfolio.temporalobs.workflows

import io.temporal.activity.ActivityInterface
import io.temporal.activity.ActivityMethod

@ActivityInterface
interface PingActivities {
    @ActivityMethod
    fun ping(): String
}
