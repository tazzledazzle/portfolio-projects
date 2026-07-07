package com.portfolio.temporalobs.workflows

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.registerKotlinModule
import io.temporal.common.converter.DataConverter
import io.temporal.common.converter.DefaultDataConverter
import io.temporal.common.converter.JacksonJsonPayloadConverter

object TemporalDataConverter {
    val instance: DataConverter by lazy {
        val mapper = ObjectMapper().registerKotlinModule()
        DefaultDataConverter.newDefaultInstance().withPayloadConverterOverrides(
            JacksonJsonPayloadConverter(mapper),
        )
    }
}
