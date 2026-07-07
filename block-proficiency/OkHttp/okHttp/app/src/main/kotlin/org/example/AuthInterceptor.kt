package org.example

import okhttp3.Interceptor
import okhttp3.Response

class AuthInterceptor(private val apiKey: String) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val authenticatedRequest = chain.request().newBuilder()
            .header("Authorization", "Bearer $apiKey")
            .header("X-Client-Version", "1.0.0")
            .build()

        return chain.proceed(authenticatedRequest)
    }
}