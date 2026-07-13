package com.marketplace.synth

import kotlin.random.Random

class Generators(private val random: Random) {
    private val nouns = listOf("Desk", "Bike", "Lamp", "Chair", "Camera", "Jacket", "Sofa", "Monitor")
    private val adjectives = listOf("Midcentury", "Vintage", "Compact", "Sturdy", "Local", "Clean")
    private val phoneLikeDigits = Regex("""\d{7,}""")

    fun buyerId(n: Int) = "synth-buyer-$n"
    fun sellerId(n: Int) = "synth-seller-$n"

    fun listingTitle(n: Int, categories: List<String>): String {
        requireCategories(categories)
        val cat = categories[n % categories.size]
        requireSafeCategory(cat)
        val adj = adjectives[n % adjectives.size]
        val noun = nouns[(n * 3) % nouns.size]
        return "Synth $adj $noun #${encodeIndex(n)} ($cat)"
    }

    fun priceCents(): Int = 1000 + random.nextInt(90_000)

    fun category(categories: List<String>): String {
        requireCategories(categories)
        val cat = categories[random.nextInt(categories.size)]
        requireSafeCategory(cat)
        return cat
    }

    /** Groups digits in threes so large indices never form a \\d{7,} run. */
    private fun encodeIndex(n: Int): String =
        n.toUInt().toString().chunked(3).joinToString("-")

    private fun requireCategories(categories: List<String>) {
        require(categories.isNotEmpty()) { "categories must not be empty" }
    }

    private fun requireSafeCategory(category: String) {
        require('@' !in category) {
            "category must not contain '@': $category"
        }
        require(!phoneLikeDigits.containsMatchIn(category)) {
            "category must not contain phone-like digit runs: $category"
        }
    }
}
