package org.example
/*
 * P1 — Deployment Dependency Scheduler
 *
 * Base:  given services and their dependencies, produce a valid deploy order.
 * Ext 1: group into parallel "waves" (everything in a wave can deploy concurrently).
 * Ext 2: detect a dependency cycle and report the actual cycle, not just "invalid".
 * Ext 3: simulate execution with K workers and per-job durations; report makespan.
 *
 * Kahn's algorithm, O(V + E). DFS only on the residual graph to name the cycle.
 */

import java.util.PriorityQueue

data class Job(
    val name: String,
    val durationSec: Int = 1,
    val dependsOn: List<String> = emptyList()
)

sealed interface ScheduleResult {
    /** waves[0] can all start immediately; waves[i] needs everything in waves[0..i-1]. */
    data class Ok(val waves: List<List<String>>) : ScheduleResult
    data class CycleDetected(val cycle: List<String>) : ScheduleResult
    data class MissingDependency(val job: String, val missing: String) : ScheduleResult
}

data class Timeline(val makespanSec: Int, val starts: Map<String, Int>, val finishes: Map<String, Int>)

class DeployScheduler(jobs: List<Job>) {

    private val byName: Map<String, Job> = jobs.associateBy { it.name }

    /** reverse edges: dependency -> jobs unblocked by it */
    private val dependents: Map<String, List<String>> =
        buildMap<String, MutableList<String>> {
            jobs.forEach { job ->
                job.dependsOn.forEach { dep -> getOrPut(dep) { mutableListOf() }.add(job.name) }
            }
        }

    init {
        require(jobs.size == byName.size) { "duplicate job names in input" }
    }

    fun plan(): ScheduleResult {
        // Validate first: an unknown dependency is a config bug, not a cycle.
        for (job in byName.values) {
            for (dep in job.dependsOn) {
                if (dep !in byName) return ScheduleResult.MissingDependency(job.name, dep)
            }
        }

        val indegree = byName.mapValues { (_, job) -> job.dependsOn.size }.toMutableMap()
        var frontier = indegree.filterValues { it == 0 }.keys.sorted()
        val waves = mutableListOf<List<String>>()
        var emitted = 0

        while (frontier.isNotEmpty()) {
            waves += frontier
            emitted += frontier.size
            val next = mutableListOf<String>()
            for (done in frontier) {
                for (child in dependents[done].orEmpty()) {
                    val remaining = indegree.getValue(child) - 1
                    indegree[child] = remaining
                    if (remaining == 0) next += child
                }
            }
            frontier = next.sorted()
        }

        if (emitted != byName.size) {
            val stuck = indegree.filterValues { it > 0 }.keys
            return ScheduleResult.CycleDetected(findCycle(stuck))
        }
        return ScheduleResult.Ok(waves)
    }

    /** Flatten waves into a single serial order. */
    fun serialOrder(): List<String>? =
        (plan() as? ScheduleResult.Ok)?.waves?.flatten()

    private fun findCycle(candidates: Set<String>): List<String> {
        val UNVISITED = 0; val ON_STACK = 1; val DONE = 2
        val state = mutableMapOf<String, Int>()
        val stack = mutableListOf<String>()

        fun dfs(node: String): List<String>? {
            state[node] = ON_STACK
            stack.add(node)
            for (dep in byName.getValue(node).dependsOn) {
                if (dep !in candidates) continue
                when (state[dep] ?: UNVISITED) {
                    UNVISITED -> dfs(dep)?.let { return it }
                    ON_STACK -> {
                        val start = stack.indexOf(dep)
                        return stack.subList(start, stack.size).toList() + dep
                    }
                }
            }
            stack.removeAt(stack.lastIndex)
            state[node] = DONE
            return null
        }

        for (node in candidates.sorted()) {
            if ((state[node] ?: UNVISITED) == UNVISITED) dfs(node)?.let { return it }
        }
        return emptyList()
    }

    /**
     * Ext 3: event-driven simulation with a fixed worker pool.
     * A job starts as soon as (a) all deps finished and (b) a worker is free.
     */
    fun simulate(workers: Int): Timeline {
        require(workers > 0) { "need at least one worker" }
        val plan = plan()
        require(plan is ScheduleResult.Ok) { "cannot simulate an invalid graph: $plan" }

        val indegree = byName.mapValues { (_, job) -> job.dependsOn.size }.toMutableMap()
        val byReadyThenName = compareBy<Pair<Int, String>>({ it.first }, { it.second })
        val ready = PriorityQueue(byReadyThenName)
        val running = PriorityQueue(byReadyThenName) // (finishAt, name)

        byName.values.filter { it.dependsOn.isEmpty() }.forEach { ready += 0 to it.name }

        val starts = mutableMapOf<String, Int>()
        val finishes = mutableMapOf<String, Int>()
        var clock = 0
        var free = workers

        while (ready.isNotEmpty() || running.isNotEmpty()) {
            while (free > 0 && ready.isNotEmpty() && ready.peek().first <= clock) {
                val name = ready.poll().second
                starts[name] = clock
                val end = clock + byName.getValue(name).durationSec
                finishes[name] = end
                running += end to name
                free--
            }
            if (running.isEmpty()) {
                clock = ready.peek().first // idle gap: fast-forward
                continue
            }
            val (finishAt, done) = running.poll()
            clock = maxOf(clock, finishAt)
            free++
            for (child in dependents[done].orEmpty()) {
                val remaining = indegree.getValue(child) - 1
                indegree[child] = remaining
                if (remaining == 0) {
                    val readyAt = byName.getValue(child).dependsOn.maxOf { finishes.getValue(it) }
                    ready += readyAt to child
                }
            }
        }
        return Timeline(finishes.values.maxOrNull() ?: 0, starts, finishes)
    }
}

// ---------------------------------------------------------------- demo / tests

private fun check(label: String, condition: Boolean) {
    if (!condition) throw AssertionError("FAILED: $label")
    println("  ok  $label")
}

fun main() {
    println("P1 — Deployment Dependency Scheduler")

    val graph = listOf(
        Job("proto", durationSec = 10),
        Job("common-lib", durationSec = 30, dependsOn = listOf("proto")),
        Job("ledger", durationSec = 60, dependsOn = listOf("common-lib")),
        Job("card-auth", durationSec = 45, dependsOn = listOf("common-lib", "proto")),
        Job("web", durationSec = 20, dependsOn = listOf("ledger", "card-auth")),
        Job("docs", durationSec = 5)
    )
    val scheduler = DeployScheduler(graph)
    val plan = scheduler.plan()
    println("waves = ${(plan as ScheduleResult.Ok).waves}")

    check("proto and docs are in wave 0", plan.waves[0] == listOf("docs", "proto"))
    check("web is last", plan.waves.last() == listOf("web"))
    val order = scheduler.serialOrder()!!
    check("serial order respects all edges",
        graph.all { job -> job.dependsOn.all { order.indexOf(it) < order.indexOf(job.name) } })

    // Ext 2 — cycle
    val cyclic = DeployScheduler(
        listOf(
            Job("a", dependsOn = listOf("c")),
            Job("b", dependsOn = listOf("a")),
            Job("c", dependsOn = listOf("b")),
            Job("standalone")
        )
    )
    val cycleResult = cyclic.plan()
    println("cycle    = ${(cycleResult as ScheduleResult.CycleDetected).cycle}")
    check("cycle has 3 distinct members", cycleResult.cycle.toSet() == setOf("a", "b", "c"))

    // Missing dependency
    val broken = DeployScheduler(listOf(Job("a", dependsOn = listOf("ghost"))))
    check("missing dep reported",
        broken.plan() == ScheduleResult.MissingDependency("a", "ghost"))

    // Ext 3 — makespan
    val unlimited = scheduler.simulate(workers = 99)
    val single = scheduler.simulate(workers = 1)
    println("makespan(∞ workers) = ${unlimited.makespanSec}s   (critical path)")
    println("makespan(1 worker)  = ${single.makespanSec}s   (sum of all work)")
    check("critical path is proto->common-lib->ledger->web = 120", unlimited.makespanSec == 120)
    check("serial total is 170", single.makespanSec == 170)
    check("2 workers sits between the two bounds",
        scheduler.simulate(2).makespanSec in 120..170)

    // Empty input
    check("empty graph is valid", (DeployScheduler(emptyList()).plan() as ScheduleResult.Ok).waves.isEmpty())

    println("P1 passed.\n")
}
