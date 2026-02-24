package showcase.core

import java.time.Instant

/**
 * Abstraction for current time to keep core free of static time and testable.
 */
interface Clock {
    fun now(): Instant
}

class SystemClock : Clock {
    override fun now(): Instant = Instant.now()
}
