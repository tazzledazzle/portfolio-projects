package com.skidroad.saga.orchestrator

import com.skidroad.saga.orchestrator.saga.SagaStateMachine
import com.skidroad.saga.proto.*
import io.grpc.Status
import io.grpc.StatusException
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import mu.KotlinLogging
import java.util.concurrent.ConcurrentHashMap

private val log = KotlinLogging.logger {}

/**
 * gRPC service implementation for the Saga Orchestrator.
 *
 * Demonstrates:
 *  - Unary RPC: [placeOrder] — request/response with full saga execution
 *  - Unary RPC: [getSagaStatus] — simple lookup
 *  - Bidirectional streaming: [streamSagaEvents] — client sends orders,
 *    server streams real-time saga state transitions back
 *
 * The saga status store is in-memory here; replace with Redis or PostgreSQL
 * for production use.
 */
class OrchestratorGrpcService(
    private val stateMachine: SagaStateMachine
) : SagaOrchestratorServiceGrpcKt.SagaOrchestratorServiceCoroutineImplBase() {

    // In-memory saga status store (keyed by saga_id)
    private val sagaStore = ConcurrentHashMap<String, SagaStatus>()

    override suspend fun placeOrder(request: PlaceOrderRequest): PlaceOrderResponse {
        log.info { "PlaceOrder received: orderId=${request.orderId}" }

        if (request.orderId.isBlank()) {
            throw StatusException(
                Status.INVALID_ARGUMENT.withDescription("order_id is required")
            )
        }

        val response = stateMachine.execute(request)

        // Persist saga status for later retrieval
        sagaStore[response.sagaStatus.sagaId] = response.sagaStatus

        return response
    }

    override suspend fun getSagaStatus(request: GetSagaStatusRequest): SagaStatus {
        return sagaStore[request.sagaId]
            ?: throw StatusException(
                Status.NOT_FOUND.withDescription("Saga not found: ${request.sagaId}")
            )
    }

    /**
     * Bidirectional streaming: each incoming [PlaceOrderRequest] triggers a saga.
     * The server emits [SagaStatus] updates as each step transitions.
     *
     * Key concept: the coroutine Flow allows us to collect inbound requests
     * lazily and emit status updates per-request without blocking.
     */
    override fun streamSagaEvents(requests: Flow<PlaceOrderRequest>): Flow<SagaStatus> = flow {
        requests.collect { request ->
            log.info { "Streaming saga for orderId=${request.orderId}" }

            // Emit PENDING immediately so the client knows we received it
            emit(sagaStatus {
                orderId = request.orderId
                state   = SagaStatus.State.PENDING
            })

            try {
                val response = stateMachine.execute(request)
                sagaStore[response.sagaStatus.sagaId] = response.sagaStatus
                emit(response.sagaStatus)
            } catch (e: StatusException) {
                // Emit FAILED status back through the stream rather than closing it
                emit(sagaStatus {
                    orderId = request.orderId
                    state   = SagaStatus.State.FAILED
                })
            }
        }
    }
}
