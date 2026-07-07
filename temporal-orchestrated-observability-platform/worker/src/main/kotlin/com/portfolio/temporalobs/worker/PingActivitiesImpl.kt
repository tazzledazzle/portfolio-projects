package com.portfolio.temporalobs.worker

import com.portfolio.temporalobs.workflows.PingActivities
import org.slf4j.LoggerFactory

class PingActivitiesImpl : PingActivities {
    override fun ping(): String {
        logger.info("PingActivity executing")
        return "pong"
    }

    companion object {
        private val logger = LoggerFactory.getLogger(PingActivitiesImpl::class.java)
    }
}
