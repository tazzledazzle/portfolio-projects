package com.marketplace.synth

import kotlinx.serialization.Serializable

@Serializable
data class Geo(val lat: Double, val lon: Double)

@Serializable
data class Profile(
    val name: String,
    val seed: Long,
    val listings: Int,
    val orders: Int,
    val confirmRatio: Double,
    val chatPairs: Int,
    val messagesPerPair: Int,
    val searchRetries: Int = 10,
    val searchRetryMs: Long = 500,
    val geo: Geo,
    val categories: List<String>
)
