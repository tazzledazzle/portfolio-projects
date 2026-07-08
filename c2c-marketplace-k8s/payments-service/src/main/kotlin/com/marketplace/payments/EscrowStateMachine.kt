package com.marketplace.payments

/**
 * Pure state machine, no I/O -- deliberately kept separate from
 * OrderRepository so it can be exhaustively unit tested (every legal and
 * illegal transition) without a database. The repository is responsible
 * for making a transition durable; this class is only responsible for
 * deciding whether a transition is legal.
 *
 * States mirror OfferUp's public "2-Day Buyer Protection": payment is held
 * on purchase, released to the seller after the buyer confirms (or the
 * protection window lapses without a dispute), refunded if the buyer
 * disputes within the window.
 */
enum class EscrowStatus { HELD, RELEASED, REFUNDED }

sealed class EscrowEvent {
    data object ConfirmDelivery : EscrowEvent()
    data object BuyerDispute : EscrowEvent()
    data object ProtectionWindowElapsed : EscrowEvent()
}

class IllegalEscrowTransitionException(message: String) : Exception(message)

object EscrowStateMachine {

    /**
     * Returns the next status for (current, event), or throws if the
     * transition isn't legal -- e.g. you can't dispute an order that's
     * already been released. Throwing here rather than returning null
     * forces every call site to handle the illegal case explicitly instead
     * of silently swallowing it.
     */
    fun transition(current: EscrowStatus, event: EscrowEvent): EscrowStatus {
        return when (current to event) {
            EscrowStatus.HELD to EscrowEvent.ConfirmDelivery -> EscrowStatus.RELEASED
            EscrowStatus.HELD to EscrowEvent.ProtectionWindowElapsed -> EscrowStatus.RELEASED
            EscrowStatus.HELD to EscrowEvent.BuyerDispute -> EscrowStatus.REFUNDED
            else -> throw IllegalEscrowTransitionException(
                "cannot apply $event to an order already in state $current"
            )
        }
    }
}
