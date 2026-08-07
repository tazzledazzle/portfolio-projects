package org.example

import java.time.*

// ============ 1. DATA MODELING ============

@JvmInline
value class ServiceId(val value: String) {
    init { require(value.isNotBlank()) { "service id must not be blank" } }
    override fun toString() = value
}

data class Deploy(
    val id: String,
    val service: ServiceId,
    val at: Instant,
    val env: Env = Env.STAGING,
    val tags: Set<String> = emptySet()
)

enum class Env(val label: String, val requiresApproval: Boolean) {
    STAGING("staging", false),
    PROD("production", true);

    companion object {
        fun from(raw: String): Env? = entries.firstOrNull { it.label.equals(raw, ignoreCase = true) }
    }
}

sealed interface Outcome<out T> {
    data class Ok<T>(val value: T) : Outcome<T>
    data class Invalid(val field: String, val reason: String) : Outcome<Nothing>
    data object NotFound : Outcome<Nothing>
}

fun <T> Outcome<T>.getOrNull(): T? = (this as? Outcome.Ok)?.value

// exhaustive when: adding a subtype breaks compilation at every use site
fun describe(o: Outcome<Deploy>): String = when (o) {
    is Outcome.Ok -> "ok ${o.value.id}"
    is Outcome.Invalid -> "invalid ${o.field}: ${o.reason}"
    Outcome.NotFound -> "not found"
}

// ============ 2. COLLECTIONS COOKBOOK ============

fun collectionsDemo(deploys: List<Deploy>) {
    val byService: Map<ServiceId, List<Deploy>> = deploys.groupBy { it.service }
    val byId: Map<String, Deploy> = deploys.associateBy { it.id }
    val envCounts: Map<Env, Int> = deploys.groupingBy { it.env }.eachCount()
    val (prod, staging) = deploys.partition { it.env == Env.PROD }
    val latestPerService = deploys.groupBy { it.service }
        .mapValues { (_, ds) -> ds.maxBy { it.at } }
    val gaps: List<Duration> = deploys.sortedBy { it.at }
        .zipWithNext { a, b -> Duration.between(a.at, b.at) }
    val rolling3: List<Int> = listOf(1, 2, 3, 4, 5).windowed(size = 3, step = 1) { it.sum() }
    val batches: List<List<Deploy>> = deploys.chunked(2)
    val allTags: Set<String> = deploys.flatMapTo(mutableSetOf()) { it.tags }
    val total = deploys.sumOf { it.tags.size }
    val index = buildMap { deploys.forEach { put(it.id, it.service) } }

    check(byService.isNotEmpty() && byId.size == deploys.size)
    check(envCounts.values.sum() == deploys.size)
    check(prod.size + staging.size == deploys.size)
    check(latestPerService.isNotEmpty())
    check(gaps.size == deploys.size - 1)
    check(rolling3 == listOf(6, 9, 12))
    check(batches.first().size == 2)
    check(allTags.isNotEmpty() && total >= 0 && index.size == deploys.size)

    // multi-key sort, nulls last
    val sorted = deploys.sortedWith(compareBy<Deploy> { it.env }.thenByDescending { it.at }.thenBy { it.id })
    check(sorted.size == deploys.size)

    // fold when you need a running accumulator, not just a sum
    val runningMax = deploys.fold(Duration.ZERO) { acc, _ -> acc.plusSeconds(1) }
    check(runningMax.seconds.toInt() == deploys.size)

    // sequence: lazy, worth it only on big inputs or early exit
    val firstProd = deploys.asSequence().filter { it.env == Env.PROD }.map { it.id }.firstOrNull()
    check(firstProd != null || prod.isEmpty())
}

// ============ 3. PARSING ============

data class ParseReporter<T>(val parsed: List<T>, val rejected: List<Pair<Int, String>>)

object LogParser {
    // "2026-08-01T09:00:00Z id=d1 service=ledger env=production tags=a,b"
    fun parse(lines: List<String>): ParseReporter<Deploy> {
        val parsed = mutableListOf<Deploy>()
        val rejected = mutableListOf<Pair<Int, String>>()

        lines.forEachIndexed { i, raw ->
            val lineNo = i + 1
            val line = raw.trim()
            if (line.isEmpty() || line.startsWith("#")) return@forEachIndexed

            val tokens = line.split(Regex("\\s+"))
            val at = runCatching { Instant.parse(tokens.first()) }.getOrElse {
                rejected += lineNo to "bad timestamp '${tokens.first()}'"; return@forEachIndexed
            }
            val fields = tokens.drop(1).mapNotNull { tok ->
                val eq = tok.indexOf('=')
                if (eq <= 0) null else tok.substring(0, eq) to tok.substring(eq + 1)
            }.toMap()

            val id = fields["id"] ?: run {
                rejected += lineNo to "missing id"; return@forEachIndexed
            }
            val svc = fields["service"] ?: run {
                rejected += lineNo to "missing service"; return@forEachIndexed
            }
            val env = Env.from(fields["env"] ?: "staging") ?: run {
                rejected += lineNo to "unknown env '${fields["env"]}'"; return@forEachIndexed
            }
            val tags = fields["tags"]?.split(",")?.filter { it.isNotBlank() }?.toSet() ?: emptySet()
            parsed += Deploy(id, ServiceId(svc), at, env, tags)
        }
        return ParseReporter(parsed.sortedBy { it.at }, rejected)
    }
}

fun safeNumbers() {
    check("42".toIntOrNull() == 42)
    check("4x2".toIntOrNull() == null)             // never toInt() on untrusted input
    check("3.5".toDoubleOrNull() == 3.5)
    check(Env.from("PRODUCTION") == Env.PROD)
    check(runCatching { Instant.parse("nope") }.isFailure)

    // regex with named groups when the format is fixed
    val re = Regex("""(?<svc>[a-z-]+)@(?<sha>[0-9a-f]{7})""")
    val m = re.matchEntire("ledger@a1b2c3d")
    check(m?.groups?.get("svc")?.value == "ledger")
    check(m?.groups?.get("sha")?.value == "a1b2c3d")

    // time arithmetic
    val t = Instant.parse("2026-08-01T09:00:00Z")
    check(Duration.between(t, t.plus(Duration.ofMinutes(90))).toMinutes() == 90L)
    check(t.atZone(ZoneOffset.UTC).dayOfWeek == DayOfWeek.SATURDAY)
    check(Duration.parse("PT2H30M").toMinutes() == 150L)
}

// ============ 4. EXTENSIBILITY: the "add two more features" patterns ============

// Pattern A — Strategy list. New requirement = new class, engine untouched.
fun interface Rule<T> { fun check(input: T): String? }

class RuleEngine<T>(private val rules: List<Rule<T>>) {
    fun violations(input: T): List<String> = rules.mapNotNull { it.check(input) }
}

// Pattern B — table-driven dispatch instead of a growing when
val PRIORITY: Map<Env, Int> = mapOf(Env.PROD to 0, Env.STAGING to 1)

// Pattern C — default params keep old call sites compiling
fun render(d: Deploy, verbose: Boolean = false, redactTags: Boolean = false): String = buildString {
    append("${d.service}/${d.id}")
    if (verbose) append(" @${d.at}")
    if (!redactTags && d.tags.isNotEmpty()) append(" [${d.tags.sorted().joinToString(",")}]")
}

// Pattern D — extension functions add behaviour without touching the type
fun Deploy.isRisky(): Boolean = env == Env.PROD && "hotfix" in tags

fun main() {
    val log = """
        # audit
        2026-08-01T09:00:00Z id=d1 service=ledger env=production tags=hotfix,urgent
        2026-08-01T09:05:00Z id=d2 service=web env=staging
        2026-08-01T09:10:00Z id=d3 service=ledger env=staging tags=nightly
        garbage line
        2026-08-01T09:15:00Z id=d4 service=web env=mars
        2026-08-01T09:20:00Z service=web env=staging
    """.trimIndent().lines()

    val report = LogParser.parse(log)
    check(report.parsed.size == 3) { "parsed ${report.parsed.size}" }
    check(report.rejected.size == 3) { "rejected ${report.rejected}" }
    println("parsed=${report.parsed.size} rejected=${report.rejected}")

    collectionsDemo(report.parsed)
    safeNumbers()

    val engine = RuleEngine<Deploy>(
        listOf(
            Rule { if (it.env == Env.PROD && it.tags.isEmpty()) "prod deploy needs a tag" else null },
            Rule { if ("hotfix" in it.tags && "approved" !in it.tags) "hotfix needs approval" else null }
        )
    )
    val v = engine.violations(report.parsed.first { it.id == "d1" })
    check(v == listOf("hotfix needs approval")) { v.toString() }

    check(render(report.parsed.first()) == "ledger/d1 [hotfix,urgent]")
    check(render(report.parsed.first(), redactTags = true) == "ledger/d1")
    check(report.parsed.first().isRisky())
    check(describe(Outcome.Ok(report.parsed.first())).startsWith("ok"))
    check(describe(Outcome.NotFound) == "not found")
    check(Outcome.Invalid("env", "unknown").getOrNull() == null)
    check(PRIORITY.getValue(Env.PROD) < PRIORITY.getValue(Env.STAGING))

    println("File A: all checks passed.")
}
