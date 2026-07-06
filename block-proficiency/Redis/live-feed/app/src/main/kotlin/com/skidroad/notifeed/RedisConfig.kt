package com.skidroad.notifeed

import io.lettuce.core.RedisClient
import io.lettuce.core.RedisURI
import io.lettuce.core.api.StatefulRedisConnection
import io.lettuce.core.pubsub.StatefulRedisPubSubConnection
import org.slf4j.LoggerFactory
import java.time.Duration

/**
 * RedisConfig — creates and owns Lettuce connections.
 *
 * Why Lettuce and not Jedis here?
 *  - Lettuce connections are thread-safe; one connection can be shared across coroutines.
 *  - Lettuce ships a native Pub/Sub API with async listeners and Kotlin coroutine bridges.
 *  - Jedis uses a blocking connection-pool model that fights against coroutines.
 *
 * Connection types used:
 *  - StatefulRedisConnection  : regular commands (LPUSH, LRANGE, LTRIM, TTL, etc.)
 *  - StatefulRedisPubSubConnection : dedicated connection for PUBLISH / SUBSCRIBE
 *    (Redis requires Pub/Sub connections to be separate from command connections)
 */
object RedisConfig {
    private val log = LoggerFactory.getLogger(RedisConfig::class.java)

    private val redisUri: RedisURI = RedisURI.builder()
        .withHost(System.getenv("REDIS_HOST") ?: "localhost")
        .withPort(System.getenv("REDIS_PORT")?.toInt() ?: 6379)
        .withTimeout(Duration.ofSeconds(5))
        .build()

    private val client: RedisClient = RedisClient.create(redisUri)

    /** General-purpose synchronous command connection (shared, thread-safe). */
    fun commandConnection(): StatefulRedisConnection<String, String> {
        log.info("Opening Redis command connection to ${redisUri.host}:${redisUri.port}")
        return client.connect()
    }

    /** Dedicated Pub/Sub connection — one per Publisher or Subscriber instance. */
    fun pubSubConnection(): StatefulRedisPubSubConnection<String, String> {
        log.info("Opening Redis Pub/Sub connection to ${redisUri.host}:${redisUri.port}")
        return client.connectPubSub()
    }

    fun shutdown() {
        log.info("Shutting down Lettuce Redis client")
        client.shutdown()
    }
}