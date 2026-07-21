package com.marketplace.synth

import kotlinx.serialization.Serializable

@Serializable
data class Summary(
    val profile: String,
    val created: Int = 0,
    val indexed: Int = 0,
    val orders: Int = 0,
    val released: Int = 0,
    val refunded: Int = 0,
    val chatOk: Boolean = false,
    val errors: List<String> = emptyList()
) {
    fun ok(): Boolean = errors.isEmpty() && created > 0
}
