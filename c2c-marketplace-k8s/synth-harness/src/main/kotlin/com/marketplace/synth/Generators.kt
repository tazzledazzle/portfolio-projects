package com.marketplace.synth

import kotlin.random.Random

class Generators(private val random: Random) {
    private val nouns = listOf("Desk", "Bike", "Lamp", "Chair", "Camera", "Jacket", "Sofa", "Monitor")
    private val adjectives = listOf("Midcentury", "Vintage", "Compact", "Sturdy", "Local", "Clean")

    fun buyerId(n: Int) = "synth-buyer-$n"
    fun sellerId(n: Int) = "synth-seller-$n"

    fun listingTitle(n: Int, categories: List<String>): String {
        val adj = adjectives[n % adjectives.size]
        val noun = nouns[(n * 3) % nouns.size]
        val cat = categories[n % categories.size]
        return "Synth $adj $noun #$n ($cat)"
    }

    fun priceCents(): Int = 1000 + random.nextInt(90_000)
    fun category(categories: List<String>) = categories[random.nextInt(categories.size)]
}
