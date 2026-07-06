package org.example

import okhttp3.Interceptor
import okhttp3.Response
import java.util.concurrent.TimeUnit

class TimingLoggingInterceptor : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val startNs = System.nanoTime()

        println("--> ${request.method} ${request.url}")
        request.headers.forEach { (name, value) ->
            // Redact sensitive headers in real code
            println("    $name: $value")
        }

        val response = chain.proceed(request)

        val elapsedMs = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startNs)
        println("<-- ${response.code} ${request.url} (${elapsedMs}ms)")

        return response
    }
}