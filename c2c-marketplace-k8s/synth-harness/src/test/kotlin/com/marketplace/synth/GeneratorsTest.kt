package com.marketplace.synth

import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test
import kotlin.random.Random

class GeneratorsTest {
    @Test
    fun `buyer and seller ids always use synth prefix`() {
        val g = Generators(Random(1))
        repeat(20) { i ->
            assertTrue(g.buyerId(i).startsWith("synth-buyer-"))
            assertTrue(g.sellerId(i).startsWith("synth-seller-"))
        }
    }

    @Test
    fun `listing titles never contain at-sign or phone-like digits runs`() {
        val g = Generators(Random(2))
        repeat(50) {
            val title = g.listingTitle(it, listOf("furniture", "electronics"))
            assertTrue('@' !in title)
            assertTrue(!Regex("""\d{7,}""").containsMatchIn(title))
            assertTrue(title.startsWith("Synth "))
        }
    }
}
