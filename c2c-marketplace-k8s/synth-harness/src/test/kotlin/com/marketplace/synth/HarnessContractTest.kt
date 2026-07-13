package com.marketplace.synth

import kotlinx.serialization.json.Json
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test
import java.io.File

class HarnessContractTest {
    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun `demo profile decodes`() {
        val text = File("synth/profiles/demo.json").readText()
        val p = json.decodeFromString<Profile>(text)
        assertEquals(10, p.listings)
        assertEquals("demo", p.name)
        assertEquals(42L, p.seed)
    }

    @Test
    fun `Summary ok requires created and no errors`() {
        assertTrue(Summary(profile = "demo", created = 1).ok())
        assertFalse(Summary(profile = "demo", created = 0).ok())
        assertFalse(Summary(profile = "demo", created = 1, errors = listOf("boom")).ok())
        assertTrue(
            Summary(profile = "demo", created = 5, chatOk = false).ok(),
            "chatOk=false must not fail ok()"
        )
    }
}
