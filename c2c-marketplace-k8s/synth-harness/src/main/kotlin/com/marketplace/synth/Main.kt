package com.marketplace.synth

import kotlinx.coroutines.runBlocking
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import java.io.File
import kotlin.system.exitProcess

fun main(args: Array<String>) {
    val profilePath = argValue(args, "--profile")
        ?: error("Usage: --profile path/to.json [--fail-fast]")
    val failFast = args.contains("--fail-fast")

    val listingsUrl = System.getenv("LISTINGS_URL") ?: "http://localhost:8081"
    val searchUrl = System.getenv("SEARCH_URL") ?: "http://localhost:8082"
    val paymentsUrl = System.getenv("PAYMENTS_URL") ?: "http://localhost:8084"

    val json = Json {
        ignoreUnknownKeys = true
        prettyPrint = true
    }

    val profile = json.decodeFromString<Profile>(File(profilePath).readText())
    val client = MarketplaceClient(listingsUrl, searchUrl, paymentsUrl)

    val summary = try {
        runBlocking {
            Harness.run(profile, client, failFast)
        }
    } finally {
        client.close()
    }

    println(json.encodeToString(summary))
    exitProcess(if (summary.ok()) 0 else 1)
}

private fun argValue(args: Array<String>, flag: String): String? {
    val idx = args.indexOf(flag)
    if (idx < 0 || idx + 1 >= args.size) return null
    return args[idx + 1]
}
