package com.skidroad.notifeed.model

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.SerializationFeature
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule
import com.fasterxml.jackson.module.kotlin.registerKotlinModule

/**
 * Shared Jackson ObjectMapper.
 * Handles Kotlin data classes + java.time.Instant serialisation.
 */
val objectMapper: ObjectMapper = ObjectMapper()
    .registerKotlinModule()
    .registerModule(JavaTimeModule())
    .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS)

fun Notification.toJson(): String = objectMapper.writeValueAsString(this)

fun String.toNotification(): Notification = objectMapper.readValue(this, Notification::class.java)