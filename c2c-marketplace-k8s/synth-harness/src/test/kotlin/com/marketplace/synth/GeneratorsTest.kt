package com.marketplace.synth

import org.junit.jupiter.api.Assertions.assertThrows
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
        val categories = listOf("furniture", "electronics")
        val indices = (0 until 50) + listOf(1_000_000, 12_345_678)
        for (n in indices) {
            val title = g.listingTitle(n, categories)
            assertTrue('@' !in title)
            assertTrue(!Regex("""\d{7,}""").containsMatchIn(title))
            assertTrue(title.startsWith("Synth "))
        }
    }

    @Test
    fun `listingTitle rejects empty categories`() {
        val g = Generators(Random(3))
        val ex = assertThrows(IllegalArgumentException::class.java) {
            g.listingTitle(0, emptyList())
        }
        assertTrue(ex.message!!.contains("categories", ignoreCase = true))
    }

    @Test
    fun `category rejects empty categories`() {
        val g = Generators(Random(4))
        val ex = assertThrows(IllegalArgumentException::class.java) {
            g.category(emptyList())
        }
        assertTrue(ex.message!!.contains("categories", ignoreCase = true))
    }

    @Test
    fun `listingTitle rejects categories with at-sign or phone-like digits`() {
        val g = Generators(Random(5))
        assertThrows(IllegalArgumentException::class.java) {
            g.listingTitle(0, listOf("bad@cat"))
        }
        assertThrows(IllegalArgumentException::class.java) {
            g.listingTitle(0, listOf("phone1234567"))
        }
    }
}
