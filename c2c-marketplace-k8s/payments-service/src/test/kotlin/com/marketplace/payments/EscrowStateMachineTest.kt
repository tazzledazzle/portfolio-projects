package com.marketplace.payments

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.Test
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.MethodSource

/**
 * Exhaustive over every (state, event) pair -- pure logic, no I/O, so
 * there's no excuse not to cover all nine combinations rather than just
 * the happy path. This is the test suite I'd actually trust to catch a
 * regression before it reaches production money-handling code.
 */
class EscrowStateMachineTest {

    @Test
    fun `confirm delivery releases a held order`() {
        assertEquals(
            EscrowStatus.RELEASED,
            EscrowStateMachine.transition(EscrowStatus.HELD, EscrowEvent.ConfirmDelivery)
        )
    }

    @Test
    fun `protection window elapsing releases a held order`() {
        assertEquals(
            EscrowStatus.RELEASED,
            EscrowStateMachine.transition(EscrowStatus.HELD, EscrowEvent.ProtectionWindowElapsed)
        )
    }

    @Test
    fun `buyer dispute refunds a held order`() {
        assertEquals(
            EscrowStatus.REFUNDED,
            EscrowStateMachine.transition(EscrowStatus.HELD, EscrowEvent.BuyerDispute)
        )
    }

    @ParameterizedTest
    @MethodSource("illegalTransitions")
    fun `illegal transitions throw rather than silently no-op`(
        current: EscrowStatus,
        event: EscrowEvent
    ) {
        assertThrows(IllegalEscrowTransitionException::class.java) {
            EscrowStateMachine.transition(current, event)
        }
    }

    companion object {
        @JvmStatic
        fun illegalTransitions(): List<Array<Any>> = listOf(
            // Can't touch an order that's already resolved, regardless of event.
            arrayOf(EscrowStatus.RELEASED, EscrowEvent.ConfirmDelivery),
            arrayOf(EscrowStatus.RELEASED, EscrowEvent.BuyerDispute),
            arrayOf(EscrowStatus.RELEASED, EscrowEvent.ProtectionWindowElapsed),
            arrayOf(EscrowStatus.REFUNDED, EscrowEvent.ConfirmDelivery),
            arrayOf(EscrowStatus.REFUNDED, EscrowEvent.BuyerDispute),
            arrayOf(EscrowStatus.REFUNDED, EscrowEvent.ProtectionWindowElapsed)
        )
    }
}
