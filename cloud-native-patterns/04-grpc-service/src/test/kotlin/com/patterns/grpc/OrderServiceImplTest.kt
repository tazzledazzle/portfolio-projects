package com.patterns.grpc

import com.patterns.grpc.proto.CreateOrderRequest
import com.patterns.grpc.proto.ListOrdersRequest
import com.patterns.grpc.proto.OrderStatus
import com.patterns.grpc.proto.OrderServiceGrpcKt
import com.patterns.grpc.proto.createOrderRequest
import com.patterns.grpc.proto.listOrdersRequest
import com.patterns.grpc.proto.orderItem
import io.grpc.ManagedChannel
import io.grpc.StatusException
import io.grpc.inprocess.InProcessChannelBuilder
import io.grpc.inprocess.InProcessServerBuilder
import io.grpc.testing.GrpcCleanupRule
import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.runBlocking
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.util.UUID

/**
 * Tests for [OrderServiceImpl] using an in-process gRPC channel.
 *
 * [InProcessChannelBuilder] creates a gRPC channel that communicates with a
 * server running in the same JVM process — no actual network socket is used.
 * This makes tests fast and hermetic without any mocking of gRPC internals.
 */
class OrderServiceImplTest {

    private val serverName = InProcessServerBuilder.generateName()
    private lateinit var channel: ManagedChannel
    private lateinit var stub: OrderServiceGrpcKt.OrderServiceCoroutineStub

    @BeforeEach
    fun setUp() {
        val server = InProcessServerBuilder
            .forName(serverName)
            .directExecutor()
            .addService(OrderServiceImpl())
            .build()
            .start()

        channel = InProcessChannelBuilder
            .forName(serverName)
            .directExecutor()
            .build()

        stub = OrderServiceGrpcKt.OrderServiceCoroutineStub(channel)

        // Register cleanup so the server and channel are shut down after each test
        Runtime.getRuntime().addShutdownHook(Thread {
            channel.shutdownNow()
            server.shutdownNow()
        })
    }

    @AfterEach
    fun tearDown() {
        channel.shutdownNow()
    }

    // ─── CreateOrder ─────────────────────────────────────────────────────────

    @Test
    fun `createOrder returns order with generated id and PENDING status`() = runBlocking {
        val response = stub.createOrder(
            createOrderRequest {
                customerId = "cust-123"
                idempotencyKey = UUID.randomUUID().toString()
                items.add(orderItem {
                    sku = "SKU-WIDGET"
                    quantity = 2
                    unitPriceCents = 999L
                })
            }
        )

        assertThat(response.order.orderId).isNotBlank()
        assertThat(response.order.customerId).isEqualTo("cust-123")
        assertThat(response.order.status).isEqualTo(OrderStatus.ORDER_STATUS_PENDING)
        assertThat(response.order.totalCents).isEqualTo(1998L) // 2 * 999
        assertThat(response.order.itemsCount).isEqualTo(1)
    }

    @Test
    fun `createOrder calculates correct total for multiple items`() = runBlocking {
        val response = stub.createOrder(
            createOrderRequest {
                customerId = "cust-456"
                idempotencyKey = UUID.randomUUID().toString()
                items.add(orderItem { sku = "A"; quantity = 3; unitPriceCents = 1000L })
                items.add(orderItem { sku = "B"; quantity = 1; unitPriceCents = 500L })
            }
        )

        assertThat(response.order.totalCents).isEqualTo(3500L)  // 3*1000 + 1*500
    }

    @Test
    fun `createOrder rejects blank customerId`() {
        assertThatThrownBy {
            runBlocking {
                stub.createOrder(
                    createOrderRequest {
                        customerId = "   "
                        idempotencyKey = UUID.randomUUID().toString()
                        items.add(orderItem { sku = "X"; quantity = 1; unitPriceCents = 100L })
                    }
                )
            }
        }.isInstanceOf(StatusException::class.java)
    }

    @Test
    fun `createOrder rejects empty items list`() {
        assertThatThrownBy {
            runBlocking {
                stub.createOrder(
                    createOrderRequest {
                        customerId = "cust-789"
                        idempotencyKey = UUID.randomUUID().toString()
                        // No items added
                    }
                )
            }
        }.isInstanceOf(StatusException::class.java)
    }

    // ─── ListOrders ──────────────────────────────────────────────────────────

    @Test
    fun `listOrders streams all orders for a customer`() = runBlocking {
        // Create 3 orders for cust-A and 1 for cust-B
        repeat(3) { i ->
            stub.createOrder(createOrderRequest {
                customerId = "cust-A"
                idempotencyKey = "key-$i"
                items.add(orderItem { sku = "SKU-$i"; quantity = 1; unitPriceCents = 100L })
            })
        }
        stub.createOrder(createOrderRequest {
            customerId = "cust-B"
            idempotencyKey = "key-cust-b"
            items.add(orderItem { sku = "SKU-B"; quantity = 1; unitPriceCents = 200L })
        })

        val ordersForA = stub.listOrders(listOrdersRequest { customerId = "cust-A" }).toList()

        assertThat(ordersForA).hasSize(3)
        assertThat(ordersForA).allMatch { it.customerId == "cust-A" }
    }

    @Test
    fun `listOrders with no customerId returns all orders`() = runBlocking {
        stub.createOrder(createOrderRequest {
            customerId = "cust-X"
            idempotencyKey = "k1"
            items.add(orderItem { sku = "S1"; quantity = 1; unitPriceCents = 50L })
        })
        stub.createOrder(createOrderRequest {
            customerId = "cust-Y"
            idempotencyKey = "k2"
            items.add(orderItem { sku = "S2"; quantity = 1; unitPriceCents = 50L })
        })

        val allOrders = stub.listOrders(listOrdersRequest { /* no customerId */ }).toList()

        assertThat(allOrders.size).isGreaterThanOrEqualTo(2)
    }

    @Test
    fun `listOrders respects maxResults limit`() = runBlocking {
        repeat(10) { i ->
            stub.createOrder(createOrderRequest {
                customerId = "cust-max"
                idempotencyKey = "max-key-$i"
                items.add(orderItem { sku = "S$i"; quantity = 1; unitPriceCents = 100L })
            })
        }

        val limited = stub.listOrders(listOrdersRequest {
            customerId = "cust-max"
            maxResults = 5
        }).toList()

        assertThat(limited).hasSize(5)
    }
}
