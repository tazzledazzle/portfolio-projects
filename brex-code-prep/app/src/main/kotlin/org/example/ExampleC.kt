package org.example


/*
 * A working back-end service with ZERO dependencies.
 * com.sun.net.httpserver.HttpServer ships with the JDK — no Ktor, no Gradle, no network.
 * This is the fastest path to "a small back-end service" in a locked-down interview sandbox.
 */

import com.sun.net.httpserver.HttpExchange
import com.sun.net.httpserver.HttpServer
import java.net.InetSocketAddress
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicInteger

// ---------- tiny JSON writer (enough for an interview; never hand-roll a parser) ----------

fun jsonEscape(s: String): String = buildString {
    s.forEach { c ->
        when (c) {
            '"' -> append("\\\""); '\\' -> append("\\\\"); '\n' -> append("\\n")
            '\r' -> append("\\r"); '\t' -> append("\\t")
            else -> if (c < ' ') append("\\u%04x".format(c.code)) else append(c)
        }
    }
}

fun toJson(value: Any?): String = when (value) {
    null -> "null"
    is String -> "\"${jsonEscape(value)}\""
    is Number, is Boolean -> value.toString()
    is Map<*, *> -> value.entries.joinToString(",", "{", "}") { "\"${jsonEscape(it.key.toString())}\":${toJson(it.value)}" }
    is Iterable<*> -> value.joinToString(",", "[", "]") { toJson(it) }
    is Enum<*> -> "\"${value.name}\""
    else -> "\"${jsonEscape(value.toString())}\""
}

// ---------- domain ----------

data class Account(val id: String, val limitCents: Long, val spentCents: Long) {
    val availableCents: Long get() = limitCents - spentCents
    fun asMap() = mapOf("id" to id, "limitCents" to limitCents, "spentCents" to spentCents, "availableCents" to availableCents)
}

class LedgerService {
    private val accounts = ConcurrentHashMap<String, Account>()
    private val idempotency = ConcurrentHashMap<String, Map<String, Any?>>()
    private val keyLocks = ConcurrentHashMap<String, Any>()

    fun seed(a: Account) { accounts[a.id] = a }
    fun get(id: String): Account? = accounts[id]

    /**
     * Idempotent by design: the same Idempotency-Key returns the same body and
     * never double-charges, even under concurrent identical requests.
     */
    fun charge(id: String, amountCents: Long, idempotencyKey: String?): Pair<Int, Map<String, Any?>> {
        if (amountCents <= 0) return 400 to mapOf("error" to "amount must be positive")
        if (idempotencyKey == null) return applyCharge(id, amountCents)

        idempotency[idempotencyKey]?.let { return 200 to it }        // fast path
        val lock = keyLocks.computeIfAbsent(idempotencyKey) { Any() }
        synchronized(lock) {
            idempotency[idempotencyKey]?.let { return 200 to it }    // re-check under lock
            try {
                val (status, body) = applyCharge(id, amountCents)
                if (status == 201) idempotency[idempotencyKey] = body // commit only on success
                return status to body
            } finally {
                keyLocks.remove(idempotencyKey, lock)
            }
        }
    }

    /** compute() holds the bin lock, so read-modify-write on one account is atomic. */
    private fun applyCharge(id: String, amountCents: Long): Pair<Int, Map<String, Any?>> {
        var status = 201
        var body: Map<String, Any?> = emptyMap()
        val found = accounts.computeIfPresent(id) { _, current ->
            if (amountCents > current.availableCents) {
                status = 409
                body = mapOf("error" to "over limit", "availableCents" to current.availableCents)
                current                                              // unchanged
            } else {
                val next = current.copy(spentCents = current.spentCents + amountCents)
                status = 201
                body = mapOf("status" to "APPROVED", "account" to next.asMap())
                next
            }
        }
        return if (found == null) 404 to mapOf("error" to "no such account: $id") else status to body
    }
}

// ---------- HTTP layer ----------

fun HttpExchange.respond(status: Int, body: Map<String, Any?>) {
    val bytes = toJson(body).toByteArray()
    responseHeaders.add("Content-Type", "application/json")
    sendResponseHeaders(status, bytes.size.toLong())
    responseBody.use { it.write(bytes) }
}

fun HttpExchange.queryParams(): Map<String, String> =
    requestURI.rawQuery.orEmpty().split("&").filter { it.isNotBlank() }
        .mapNotNull { p -> p.split("=", limit = 2).takeIf { it.size == 2 }?.let { it[0] to it[1] } }
        .toMap()

fun startServer(service: LedgerService, port: Int = 0): HttpServer {
    val server = HttpServer.create(InetSocketAddress(port), 0)
    server.executor = Executors.newFixedThreadPool(4)

    server.createContext("/accounts") { ex ->
        runCatching {
            val id = ex.requestURI.path.removePrefix("/accounts/").substringBefore("/")
            when (ex.requestMethod) {
                "GET" -> service.get(id)
                    ?.let { ex.respond(200, it.asMap()) }
                    ?: ex.respond(404, mapOf("error" to "not found"))

                "POST" -> {
                    val amount = ex.queryParams()["amount"]?.toLongOrNull()
                        ?: return@runCatching ex.respond(400, mapOf("error" to "amount required"))
                    val key = ex.requestHeaders.getFirst("Idempotency-Key")
                    val (status, body) = service.charge(id, amount, key)
                    ex.respond(status, body)
                }

                else -> ex.respond(405, mapOf("error" to "method not allowed"))
            }
        }.onFailure { ex.respond(500, mapOf("error" to it.message.orEmpty())) }
    }

    server.start()
    return server
}

// ---------- client side: java.net.http, also zero-dependency ----------

fun main() {
    val service = LedgerService().apply { seed(Account("acct-1", limitCents = 10_000, spentCents = 0)) }
    val server = startServer(service)
    val base = "http://localhost:${server.address.port}"
    val client = HttpClient.newHttpClient()

    fun get(path: String): HttpResponse<String> = client.send(
        HttpRequest.newBuilder(URI.create("$base$path")).GET().build(),
        HttpResponse.BodyHandlers.ofString()
    )

    fun post(path: String, key: String? = null): HttpResponse<String> {
        val b = HttpRequest.newBuilder(URI.create("$base$path")).POST(HttpRequest.BodyPublishers.noBody())
        key?.let { b.header("Idempotency-Key", it) }
        return client.send(b.build(), HttpResponse.BodyHandlers.ofString())
    }

    try {
        val initial = get("/accounts/acct-1")
        println("GET  -> ${initial.statusCode()} ${initial.body()}")
        check(initial.statusCode() == 200 && "10000" in initial.body())

        check(get("/accounts/nope").statusCode() == 404)
        check(post("/accounts/acct-1").statusCode() == 400)          // missing amount

        val charge = post("/accounts/acct-1?amount=3000", key = "req-1")
        println("POST -> ${charge.statusCode()} ${charge.body()}")
        check(charge.statusCode() == 201 && "APPROVED" in charge.body())

        val replay = post("/accounts/acct-1?amount=3000", key = "req-1")
        println("replay -> ${replay.statusCode()} ${replay.body()}")
        check(replay.statusCode() == 200)
        check(service.get("acct-1")!!.spentCents == 3000L) { "double charged: ${service.get("acct-1")}" }

        check(post("/accounts/acct-1?amount=99999", key = "req-2").statusCode() == 409)

        // concurrent identical requests must not double-charge
        val hits = AtomicInteger(0)
        val pool = Executors.newFixedThreadPool(8)
        val before = service.get("acct-1")!!.spentCents
        repeat(8) { pool.submit { post("/accounts/acct-1?amount=100", key = "req-3"); hits.incrementAndGet() } }
        pool.shutdown(); pool.awaitTermination(20, java.util.concurrent.TimeUnit.SECONDS)
        val after = service.get("acct-1")!!.spentCents
        println("8 concurrent identical charges: spent $before -> $after (all 8 got a response: ${hits.get() == 8})")
        check(after - before == 100L) { "double charged: $before -> $after" }

        check(toJson(mapOf("a" to listOf(1, 2), "b" to "he said \"hi\"")) == """{"a":[1,2],"b":"he said \"hi\""}""")

        println("File C: all checks passed.")
    } finally {
        // Gotcha: stop() does NOT shut down an executor you supplied, and its
        // threads are non-daemon — the JVM will hang forever without this.
        server.stop(0)
        (server.executor as java.util.concurrent.ExecutorService).shutdownNow()
    }
}
