/*
 * P3 - Deploy Gate / Rollout Policy Engine   (the "Craft" format: base + extensions)
 *
 * Base:  canDeploy(request) -> allow or deny, given required CI checks.
 * Ext 1: multiple rules with precedence, and an explanation of every violation.
 * Ext 2: time-based change-freeze windows.
 * Ext 3: break-glass override that downgrades a hard deny to "needs approval",
 *        plus environment promotion ordering (staging before prod).
 *
 * The shape that scores here is an open Rule interface: each new requirement is a
 * new class, not another branch in a growing if-tree.
 */

import java.time.DayOfWeek
import java.time.Instant
import java.time.ZoneOffset

data class DeployRequest(
    val service: String,
    val environment: String,          // "staging" | "production"
    val at: Instant,
    val checks: Map<String, Boolean>, // e.g. "unit-tests" to true
    val actor: String,
    val labels: Set<String> = emptySet(),
    /** environments this exact artifact is already live in */
    val promotedFrom: Set<String> = emptySet()
)

enum class Verdict { ALLOW, NEEDS_APPROVAL, DENY }

data class Violation(val ruleId: String, val verdict: Verdict, val reason: String)

data class GateDecision(val verdict: Verdict, val violations: List<Violation>) {
    val allowed: Boolean get() = verdict == Verdict.ALLOW
    override fun toString(): String =
        if (violations.isEmpty()) "ALLOW"
        else "$verdict <- " + violations.joinToString("; ") { "[${it.ruleId}] ${it.reason}" }
}

fun interface Rule {
    /** null means "this rule has no objection". */
    fun evaluate(request: DeployRequest): Violation?
}

class RequiredChecksRule(private val required: Set<String>) : Rule {
    override fun evaluate(request: DeployRequest): Violation? {
        val missing = required.filter { request.checks[it] != true }.sorted()
        return if (missing.isEmpty()) null
        else Violation("required-checks", Verdict.DENY, "checks not green: ${missing.joinToString(",")}")
    }
}

/** Ext 2: no production deploys Friday 17:00 UTC through Monday 09:00 UTC. */
class ChangeFreezeRule(private val environments: Set<String> = setOf("production")) : Rule {
    override fun evaluate(request: DeployRequest): Violation? {
        if (request.environment !in environments) return null
        val t = request.at.atZone(ZoneOffset.UTC)
        val frozen = when (t.dayOfWeek) {
            DayOfWeek.FRIDAY -> t.hour >= 17
            DayOfWeek.SATURDAY, DayOfWeek.SUNDAY -> true
            DayOfWeek.MONDAY -> t.hour < 9
            else -> false
        }
        return if (!frozen) null
        else Violation("change-freeze", Verdict.DENY, "weekend freeze window is active")
    }
}

/** Ext 3: artifact must be live in staging before it can go to production. */
class PromotionOrderRule(private val ladder: List<String> = listOf("staging", "production")) : Rule {
    override fun evaluate(request: DeployRequest): Violation? {
        val idx = ladder.indexOf(request.environment)
        if (idx <= 0) return null
        val prerequisite = ladder[idx - 1]
        return if (prerequisite in request.promotedFrom) null
        else Violation("promotion-order", Verdict.DENY, "not yet promoted from $prerequisite")
    }
}

class ActorAuthorizationRule(private val allowed: Set<String>) : Rule {
    override fun evaluate(request: DeployRequest): Violation? =
        if (request.actor in allowed) null
        else Violation("actor-auth", Verdict.NEEDS_APPROVAL, "${request.actor} is not a release owner")
}

class DeployGate(
    private val rules: List<Rule>,
    /** label that softens hard denials into an approval request */
    private val breakGlassLabel: String = "break-glass",
    /** rules that break-glass may never override */
    private val nonOverridable: Set<String> = setOf("required-checks")
) {
    fun evaluate(request: DeployRequest): GateDecision {
        val raw = rules.mapNotNull { it.evaluate(request) }
        if (raw.isEmpty()) return GateDecision(Verdict.ALLOW, emptyList())

        val breakGlass = breakGlassLabel in request.labels
        val effective = raw.map { v ->
            if (breakGlass && v.verdict == Verdict.DENY && v.ruleId !in nonOverridable)
                v.copy(verdict = Verdict.NEEDS_APPROVAL, reason = v.reason + " (break-glass requested)")
            else v
        }
        // Report every violation, but the overall verdict is the strictest one.
        val worst = if (effective.any { it.verdict == Verdict.DENY }) Verdict.DENY else Verdict.NEEDS_APPROVAL
        return GateDecision(worst, effective)
    }
}

// ---------------------------------------------------------------- demo / tests

private fun check(label: String, condition: Boolean) {
    if (!condition) throw AssertionError("FAILED: $label")
    println("  ok  $label")
}

fun main() {
    println("P3 - Deploy Gate / Rollout Policy Engine")

    val gate = DeployGate(
        listOf(
            RequiredChecksRule(setOf("unit-tests", "integration-tests", "security-scan")),
            ChangeFreezeRule(),
            PromotionOrderRule(),
            ActorAuthorizationRule(setOf("terence", "release-bot"))
        )
    )

    val green = mapOf("unit-tests" to true, "integration-tests" to true, "security-scan" to true)
    val tuesday = Instant.parse("2026-08-04T15:00:00Z")   // Tuesday
    val friday = Instant.parse("2026-08-07T18:00:00Z")    // Friday evening

    val clean = DeployRequest("ledger", "production", tuesday, green, "terence", promotedFrom = setOf("staging"))
    println("clean        -> ${gate.evaluate(clean)}")
    check("clean request allowed", gate.evaluate(clean).allowed)

    val redCi = clean.copy(checks = green + ("security-scan" to false))
    println("red CI       -> ${gate.evaluate(redCi)}")
    check("red CI denied", gate.evaluate(redCi).verdict == Verdict.DENY)
    check("red CI names the failing check",
        gate.evaluate(redCi).violations.single().reason.contains("security-scan"))

    val frozen = clean.copy(at = friday)
    println("freeze       -> ${gate.evaluate(frozen)}")
    check("freeze window denies", gate.evaluate(frozen).verdict == Verdict.DENY)

    val unpromoted = clean.copy(promotedFrom = emptySet())
    check("unpromoted artifact denied", gate.evaluate(unpromoted).verdict == Verdict.DENY)
    check("staging needs no promotion", gate.evaluate(clean.copy(environment = "staging")).allowed)

    val stranger = clean.copy(actor = "intern")
    println("stranger     -> ${gate.evaluate(stranger)}")
    check("unknown actor needs approval", gate.evaluate(stranger).verdict == Verdict.NEEDS_APPROVAL)

    // Ext 3 - break-glass
    val breakGlass = frozen.copy(labels = setOf("break-glass"))
    println("break-glass  -> ${gate.evaluate(breakGlass)}")
    check("break-glass downgrades the freeze", gate.evaluate(breakGlass).verdict == Verdict.NEEDS_APPROVAL)

    val breakGlassRedCi = redCi.copy(labels = setOf("break-glass"))
    println("bg + red CI  -> ${gate.evaluate(breakGlassRedCi)}")
    check("break-glass cannot override red CI",
        gate.evaluate(breakGlassRedCi).verdict == Verdict.DENY)

    // Multiple simultaneous violations are all reported, not just the first.
    val messy = DeployRequest("ledger", "production", friday, green + ("unit-tests" to false), "intern")
    val decision = gate.evaluate(messy)
    println("messy        -> $decision")
    check("all four rules fire", decision.violations.size == 4)
    check("messy is denied", decision.verdict == Verdict.DENY)

    check("a gate with no rules allows everything",
        DeployGate(emptyList()).evaluate(messy).allowed)

    println("P3 passed.\n")
}
