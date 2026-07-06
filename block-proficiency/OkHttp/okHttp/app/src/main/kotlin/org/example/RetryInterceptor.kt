package org.example

import okhttp3.Interceptor
import okhttp3.Response
import java.io.IOException

class RetryInterceptor(
    private val maxRetries: Int = 3,
    private val retryOnCodes: Set<Int> = setOf(429, 503)
) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        var response: Response? = null
        var attempt = 0
        var lastException: IOException? = null

        while (attempt <= maxRetries) {
            try {
                response?.close()                      // close previous failed response body
                response = chain.proceed(request)

                if (response.isSuccessful || response.code !in retryOnCodes) {
                    return response                    // success or non-retryable code → done
                }

                val retryAfterMs = response.header("Retry-After")
                    ?.toLongOrNull()
                    ?.times(1000)
                    ?: backoffMs(attempt)

                println("Retry ${attempt + 1}/$maxRetries after ${retryAfterMs}ms (HTTP ${response.code})")
                Thread.sleep(retryAfterMs)

            } catch (e: IOException) {
                lastException = e
                if (attempt == maxRetries) break
                Thread.sleep(backoffMs(attempt))
            }
            attempt++
        }

        return response ?: throw lastException ?: IOException("Max retries ($maxRetries) exceeded")
    }

    // 1s, 2s, 4s, 8s … capped at 30s
    private fun backoffMs(attempt: Int): Long =
        minOf(1000L * (1L shl attempt), 30_000L)
}