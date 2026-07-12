# Synthetic Data Factory Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Ship a reviewable synthetic-data harness plus Cursor subagent that safely floods the local C2C marketplace (listings → search → purchase → chat) using `synth-*` identities only.

**Architecture:** Hybrid — Kotlin `:synth-harness` owns HTTP generation/assertions; `scripts/synth-chat.sh` drives WebSocket peers; `scripts/synth-run.sh` selects a JSON profile; Cursor agent `synth-data-factory` only orchestrates (never invents PII).

**Tech Stack:** Kotlin 1.9 / JVM 17, Ktor Client 2.3.12, kotlinx.serialization, JUnit 5, Gradle multi-module, bash + websocat (or ktor-client-websockets), kind/compose localhost ports 8081–8084.

**Design:** `docs/plans/2026-07-12-synth-data-factory-design.md`

---

### Task 1: Scaffold `:synth-harness` module + profiles

**Files:**
- Modify: `settings.gradle.kts`
- Create: `synth-harness/build.gradle.kts`
- Create: `synth/profiles/demo.json`
- Create: `synth/profiles/load-light.json`
- Create: `synth-harness/src/main/kotlin/com/marketplace/synth/Profile.kt`

**Step 1: Register the module**

In `settings.gradle.kts`, change include line to:

```kotlin
include(":common", ":listings-service", ":search-service", ":messaging-service", ":payments-service", ":synth-harness")
```

**Step 2: Add `synth-harness/build.gradle.kts`**

```kotlin
plugins {
    kotlin("jvm")
    kotlin("plugin.serialization")
    application
}

application {
    mainClass.set("com.marketplace.synth.MainKt")
}

dependencies {
    val ktorVersion = "2.3.12"
    implementation("io.ktor:ktor-client-core:$ktorVersion")
    implementation("io.ktor:ktor-client-cio:$ktorVersion")
    implementation("io.ktor:ktor-client-content-negotiation:$ktorVersion")
    implementation("io.ktor:ktor-serialization-kotlinx-json:$ktorVersion")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.3")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.1")
    implementation("ch.qos.logback:logback-classic:1.5.6")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.2")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

kotlin {
    jvmToolchain(17)
}

tasks.test {
    useJUnitPlatform()
}
```

**Step 3: Create profiles**

`synth/profiles/demo.json`:

```json
{
  "name": "demo",
  "seed": 42,
  "listings": 10,
  "orders": 5,
  "confirmRatio": 0.6,
  "chatPairs": 1,
  "messagesPerPair": 3,
  "searchRetries": 10,
  "searchRetryMs": 500,
  "geo": { "lat": 47.6062, "lon": -122.3321 },
  "categories": ["furniture", "sporting-goods", "electronics"]
}
```

`synth/profiles/load-light.json`: same shape with `"listings": 100`, `"orders": 20`, `"chatPairs": 3`, `"messagesPerPair": 5`, `"seed": 99`.

**Step 4: Profile data class**

```kotlin
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
```

**Step 5: Verify module resolves**

```bash
./gradlew :synth-harness:compileKotlin
```

Expected: `BUILD SUCCESSFUL`

**Step 6: Commit**

```bash
git add settings.gradle.kts synth-harness/build.gradle.kts synth/profiles synth-harness/src/main/kotlin/com/marketplace/synth/Profile.kt
git commit -m "feat(synth): scaffold synth-harness module and demo/load-light profiles"
```

---

### Task 2: TDD generators (safe-by-construction IDs/titles)

**Files:**
- Create: `synth-harness/src/main/kotlin/com/marketplace/synth/Generators.kt`
- Create: `synth-harness/src/test/kotlin/com/marketplace/synth/GeneratorsTest.kt`

**Step 1: Write failing tests**

```kotlin
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
```

**Step 2: Run tests — expect fail**

```bash
./gradlew :synth-harness:test --tests "com.marketplace.synth.GeneratorsTest"
```

Expected: compile failure / class not found for `Generators`

**Step 3: Minimal implementation**

```kotlin
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
```

**Step 4: Run tests — expect pass**

```bash
./gradlew :synth-harness:test --tests "com.marketplace.synth.GeneratorsTest"
```

Expected: `BUILD SUCCESSFUL`, 2 tests passed

**Step 5: Commit**

```bash
git add synth-harness/src/main/kotlin/com/marketplace/synth/Generators.kt \
        synth-harness/src/test/kotlin/com/marketplace/synth/GeneratorsTest.kt
git commit -m "feat(synth): add safe-by-construction ID and title generators"
```

---

### Task 3: HTTP client + listing/search/order flows with assertions

**Files:**
- Create: `synth-harness/src/main/kotlin/com/marketplace/synth/MarketplaceClient.kt`
- Create: `synth-harness/src/main/kotlin/com/marketplace/synth/Harness.kt`
- Create: `synth-harness/src/main/kotlin/com/marketplace/synth/Main.kt`
- Create: `synth-harness/src/main/kotlin/com/marketplace/synth/Summary.kt`
- Create: `synth-harness/src/test/kotlin/com/marketplace/synth/HarnessContractTest.kt` (optional WireMock-free: test generators+profile load only if no mock server; prefer testing `Summary` merge + profile decode)

**Step 1: Summary + DTOs**

```kotlin
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
```

**Step 2: MarketplaceClient** (ktor client)

Implement methods:
- `createListing(...)` → POST `$listingsUrl/listings`
- `search(q, lat, lon)` → GET `$searchUrl/search?...`
- `createOrder(...)` → POST `$paymentsUrl/orders`
- `confirmDelivery(id)` / `dispute(id)`

Use `ContentNegotiation { json() }`. URLs from constructor.

**Step 3: Harness.run(profile, urls): Summary**

Algorithm:
1. Create `profile.listings` listings with `Generators(Random(profile.seed))`.
2. For each listing, retry search for a distinctive token from the title until found or retries exhausted; count `indexed` / push error.
3. Create up to `profile.orders` orders pairing synth buyer/seller; with probability `confirmRatio` call confirm-delivery else dispute; count released/refunded.
4. Return Summary (chatOk left false — filled by runner after chat script).

**Step 4: Main.kt**

```kotlin
fun main(args: Array<String>) {
    // args: --profile path/to.json [--fail-fast]
    // env: LISTINGS_URL SEARCH_URL PAYMENTS_URL (defaults localhost 8081/8082/8084)
}
```

Print summary JSON to stdout; `exitProcess(0/1)` based on `summary.ok()` (chat ignored here).

**Step 5: Unit test profile decode**

```kotlin
@Test
fun `demo profile decodes`() {
    val text = java.io.File("synth/profiles/demo.json").readText()
    val p = Json.decodeFromString<Profile>(text)
    assertEquals(10, p.listings)
}
```

Run from repo root working dir — configure test workingDir in Gradle if needed:

```kotlin
tasks.test {
    useJUnitPlatform()
    workingDir = rootProject.projectDir
}
```

**Step 6: Commit**

```bash
git add synth-harness/src
git commit -m "feat(synth): HTTP harness for listings, search wait, and escrow orders"
```

---

### Task 4: Chat driver script

**Files:**
- Create: `scripts/synth-chat.sh`

**Prerequisites:** `websocat` installed (`brew install websocat`) OR document fallback.

**Step 1: Write script**

Behavior:
- Args: `--url ws://localhost:8083` `--buyer synth-buyer-0` `--seller synth-seller-0` `--messages 3`
- Background `websocat` for seller reading lines; buyer sends JSON `{"conversationId":"buyer:seller","body":"Synth hello N"}`
- Assert seller stdout contains at least one `Synth hello`
- Exit 0/1

Use conversationId `"$buyer:$seller"`.

**Step 2: Make executable**

```bash
chmod +x scripts/synth-chat.sh
```

**Step 3: Manual smoke** (cluster must be up)

```bash
./scripts/synth-chat.sh --buyer synth-buyer-0 --seller synth-seller-0 --messages 2
```

Expected: exit 0

**Step 4: Commit**

```bash
git add scripts/synth-chat.sh
git commit -m "feat(synth): add WebSocket chat driver for synthetic peers"
```

---

### Task 5: `synth-run.sh` runner

**Files:**
- Create: `scripts/synth-run.sh`

**Step 1: Implement**

```bash
#!/usr/bin/env bash
set -euo pipefail
PROFILE_NAME="${1:-demo}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export LISTINGS_URL="${LISTINGS_URL:-http://localhost:8081}"
export SEARCH_URL="${SEARCH_URL:-http://localhost:8082}"
export MESSAGING_WS_URL="${MESSAGING_WS_URL:-ws://localhost:8083}"
export PAYMENTS_URL="${PAYMENTS_URL:-http://localhost:8084}"

# health checks on /healthz for listings+payments+search
# run: ./gradlew :synth-harness:run --args="--profile $ROOT/synth/profiles/${PROFILE_NAME}.json"
# capture JSON summary
# run synth-chat.sh for chatPairs from profile (parse with jq)
# merge chatOk into final summary; exit accordingly
```

**Step 2: chmod +x**

**Step 3: Live verify on kind/compose**

```bash
./scripts/synth-run.sh demo
```

Expected: prints summary JSON and `PASS` / exit 0

**Step 4: Commit**

```bash
git add scripts/synth-run.sh
git commit -m "feat(synth): add synth-run.sh profile runner"
```

---

### Task 6: Cursor subagent definition

**Files:**
- Create: `.claude/agents/synth-data-factory.md`  
  (If project uses `~/.claude/agents`, also document copy path in README snippet.)

**Step 1: Write agent markdown**

Must include:
- Role: orchestrator only
- Hard rules: never invent emails/phones/real names; always use harness; only localhost/kind NodePorts by convention
- Steps: confirm services healthy → `./scripts/synth-run.sh <profile>` → return summary
- Allowed profiles: `demo`, `load-light`
- On failure: paste harness stderr + summary.errors

**Step 2: Commit**

```bash
git add .claude/agents/synth-data-factory.md
git commit -m "feat(synth): add Cursor synth-data-factory subagent"
```

---

### Task 7: README + design checklist verification

**Files:**
- Modify: `README.md` (short “Synthetic data” section)
- Modify: `docs/plans/2026-07-12-synth-data-factory-design.md` success criteria checkboxes if desired

**Step 1: Document commands**

```bash
./scripts/synth-run.sh demo
./scripts/synth-run.sh load-light
```

**Step 2: Run verification checklist**

- [ ] `./gradlew :synth-harness:test`
- [ ] `./scripts/synth-run.sh demo` against kind → PASS
- [ ] Subagent instructions reviewed (no PII invention)

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: document synthetic data factory usage"
```

---

## Verification Checklist

- [ ] Generators tests green
- [ ] `demo` profile greens on kind (listings indexed, orders settle, chat ok)
- [ ] `load-light` completes without service crash
- [ ] Subagent file present and orchestration-only
- [ ] No direct DB writes in harness

## Rollback

Stop `synth-run.sh` / Gradle run. Recreate kind cluster or truncate service DBs if a clean slate is needed. Synthetic rows are identifiable by `synth-` prefixes.
