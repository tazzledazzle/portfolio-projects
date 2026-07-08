package com.patterns.cqrs

import com.patterns.cqrs.aggregate.InventoryAggregate
import com.patterns.cqrs.commands.Command
import com.patterns.cqrs.commands.DeductInventoryCommand
import com.patterns.cqrs.commands.InitializeInventoryCommand
import com.patterns.cqrs.commands.RestockInventoryCommand
import com.patterns.cqrs.events.DomainEvent
import com.patterns.cqrs.projection.InventoryProjection
import org.slf4j.LoggerFactory

/**
 * Routes commands to their aggregate handlers and persists resulting events.
 *
 * The flow for each command:
 *   1. Load the aggregate from the EventStore (replay from snapshot + delta events)
 *   2. Call the appropriate handle() method on the aggregate
 *   3. Append the resulting events to the EventStore
 *   4. Update the read projection(s) with the new events
 *
 * In production with Axon Framework, this is wired by annotations (@CommandHandler,
 * @EventHandler). Here it's explicit to show the mechanics.
 */
class CommandBus(
    private val eventStore: EventStore,
    private val projection: InventoryProjection,
) {
    private val log = LoggerFactory.getLogger(CommandBus::class.java)

    fun dispatch(command: Command): List<DomainEvent> {
        log.info("[CommandBus] Dispatching {} for aggregate={}", command::class.simpleName, command.aggregateId)

        val events: List<DomainEvent> = when (command) {
            is InitializeInventoryCommand -> handleInitialize(command)
            is DeductInventoryCommand -> handleDeduct(command)
            is RestockInventoryCommand -> handleRestock(command)
        }

        // Update read projections synchronously in this demo.
        // In production, projections are often updated asynchronously from Kafka events.
        events.forEach { projection.on(it) }

        log.info(
            "[CommandBus] {} produced {} event(s): {}",
            command::class.simpleName,
            events.size,
            events.map { it::class.simpleName },
        )

        return events
    }

    private fun handleInitialize(cmd: InitializeInventoryCommand): List<DomainEvent> {
        val aggregate = InventoryAggregate(cmd.aggregateId)
        aggregate.handle(cmd)
        val events = aggregate.drainPendingEvents()
        eventStore.append(cmd.aggregateId, events, aggregate)
        return events
    }

    private fun handleDeduct(cmd: DeductInventoryCommand): List<DomainEvent> {
        val aggregate = eventStore.load(cmd.aggregateId)
        val events = aggregate.handle(cmd)  // handle() drains pending internally
        eventStore.append(cmd.aggregateId, events, aggregate)
        return events
    }

    private fun handleRestock(cmd: RestockInventoryCommand): List<DomainEvent> {
        val aggregate = eventStore.load(cmd.aggregateId)
        aggregate.handle(cmd)
        val events = aggregate.drainPendingEvents()
        eventStore.append(cmd.aggregateId, events, aggregate)
        return events
    }
}
