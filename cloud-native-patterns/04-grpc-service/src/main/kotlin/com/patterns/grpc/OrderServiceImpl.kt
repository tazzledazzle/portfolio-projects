package com.patterns.grpc

import com.patterns.grpc.proto.CreateOrderRequest
import com.patterns.grpc.proto.CreateOrderResponse
import com.patterns.grpc.proto.ListOrdersRequest
import com.patterns.grpc.proto.Order
import com.patterns.grpc.proto.OrderItem
import com.patterns.grpc.proto.OrderServiceGrpcKt
import com.patterns.grpc.proto.OrderStatus
import com.patterns.grpc.proto.order
import com.patterns.grpc.proto.createOrderResponse
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.asFlow
import kotlinx.coroutines.flow.filter
import org.slf4j.LoggerFactory
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap

/**
 * In-memory order store for the demo. In production this would delegate to
 * a repository backed by a database.
 */
class OrderStore {
    private val orders = ConcurrentHashMap<String, Order>()

    fun save(order: Order): Order {
        orders[order.orderId] = order
        return order
    }

    fun findAll(): List<Order> = orders.values.toList()

    fun findByCustomer(customerId: String): List<Order> =
        orders.values.filter { it.customerId == customerId }

    fun findById(orderId: String): Order? = orders[orderId]
}

/**
 * Coroutine-based implementation of [OrderServiceGrpcKt.OrderServiceCoroutineImplBase].
 *
 * gRPC-Kotlin generates coroutine stubs — `suspend fun` for unary RPCs and
 * `Flow<T>` for server-streaming RPCs. No callbacks or ListenableFutures needed.
 */
class OrderServiceImpl(
    private val store: OrderStore = OrderStore(),
) : OrderServiceGrpcKt.OrderServiceCoroutineImplBase() {

    private val log = LoggerFactory.getLogger(OrderServiceImpl::class.java)

    /**
     * Unary RPC: create a single order and return it.
     *
     * Idempotency: in production, look up [CreateOrderRequest.idempotencyKey] in a
     * cache/DB before creating. If found, return the existing order. This ensures
     * safe retries without creating duplicates.
     */
    override suspend fun createOrder(request: CreateOrderRequest): CreateOrderResponse {
        log.info(
            "CreateOrder customerId={} items={} idempotencyKey={}",
            request.customerId,
            request.itemsCount,
            request.idempotencyKey,
        )
        require(request.customerId.isNotBlank()) { "customerId must not be blank" }
        require(request.itemsList.isNotEmpty()) { "Order must have at least one item" }

        val totalCents = request.itemsList.sumOf { it.unitPriceCents * it.quantity }
        val newOrder = order {
            orderId = UUID.randomUUID().toString()
            customerId = request.customerId
            items.addAll(request.itemsList)
            status = OrderStatus.ORDER_STATUS_PENDING
            createdAtUnixMs = System.currentTimeMillis()
            this.totalCents = totalCents
        }

        store.save(newOrder)
        log.info("Order created orderId={} totalCents={}", newOrder.orderId, totalCents)

        return createOrderResponse { order = newOrder }
    }

    /**
     * Server-streaming RPC: stream orders matching the filter back to the client.
     *
     * Returns a [Flow] — gRPC-Kotlin sends each emitted value as a separate
     * streaming message. The client receives them incrementally without waiting
     * for the full result set.
     */
    override fun listOrders(request: ListOrdersRequest): Flow<Order> {
        log.info(
            "ListOrders customerId='{}' statusFilter={} maxResults={}",
            request.customerId,
            request.statusFilter,
            request.maxResults,
        )

        val candidates = if (request.customerId.isNotBlank()) {
            store.findByCustomer(request.customerId)
        } else {
            store.findAll()
        }

        val statusFilter = request.statusFilter
        val maxResults = if (request.maxResults > 0) request.maxResults else 100

        return candidates
            .filter { order ->
                statusFilter == OrderStatus.ORDER_STATUS_UNKNOWN || order.status == statusFilter
            }
            .take(maxResults)
            .asFlow()
    }
}
