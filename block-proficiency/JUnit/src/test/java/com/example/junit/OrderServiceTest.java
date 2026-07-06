package com.example.junit;

import org.junit.jupiter.api.*;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.*;
import org.mockito.*;
import org.mockito.junit.jupiter.MockitoExtension;
import org.junit.jupiter.api.extension.ExtendWith;

import java.math.BigDecimal;
import java.util.List;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderServiceTest {

    // ── Mockito wiring ────────────────────────────────────────────────────
    @Mock
    private OrderRepository repository;

    @Mock
    private PaymentGateway paymentGateway;

    @InjectMocks
    private OrderService orderService;   // constructor-injected with the two mocks above

    // ── Lifecycle ─────────────────────────────────────────────────────────
    @BeforeAll
    static void initSuite() {
        // Runs once before any test in this class — good for expensive shared setup
        System.out.println("Suite starting");
    }

    @AfterAll
    static void tearDownSuite() {
        System.out.println("Suite done");
    }

    @BeforeEach
    void setUp() {
        // Runs before each test — reset any shared state
    }

    @AfterEach
    void tearDown() {
        // Runs after each test — release per-test resources
    }

    // ── Basic test ────────────────────────────────────────────────────────
    @Test
    void placeOrder_persistsAndCharges() {
        var order = new Order("cust-1", new BigDecimal("49.99"));
        when(repository.save(order)).thenReturn(order.withId(42L));
        when(paymentGateway.charge("cust-1", new BigDecimal("49.99"))).thenReturn("txn-abc");

        Receipt receipt = orderService.placeOrder(order);

        assertAll(
            () -> assertEquals(42L,      receipt.orderId()),
            () -> assertEquals("txn-abc", receipt.transactionId()),
            () -> assertNotNull(receipt.placedAt())
        );

        // verify exact interactions
        verify(repository, times(1)).save(order);
        verify(paymentGateway, times(1)).charge("cust-1", new BigDecimal("49.99"));
    }

    // ── assertThrows ──────────────────────────────────────────────────────
    @Test
    void placeOrder_throwsWhenAmountNegative() {
        var bad = new Order("cust-1", new BigDecimal("-1.00"));

        assertThrows(IllegalArgumentException.class, () -> orderService.placeOrder(bad));

        verifyNoInteractions(repository, paymentGateway);
    }

    // ── assertTimeout ─────────────────────────────────────────────────────
    @Test
    void placeOrder_completesWithinTimeLimit() {
        var order = new Order("cust-1", new BigDecimal("10.00"));
        when(repository.save(order)).thenReturn(order.withId(1L));
        when(paymentGateway.charge(any(), any())).thenReturn("txn-fast");

        assertTimeout(java.time.Duration.ofMillis(200), () -> orderService.placeOrder(order));
    }

    // ── @ValueSource ──────────────────────────────────────────────────────
    @ParameterizedTest
    @ValueSource(strings = {"", "  ", "\t"})
    void placeOrder_throwsWhenCustomerIdBlank(String customerId) {
        var order = new Order(customerId, new BigDecimal("10.00"));
        assertThrows(IllegalArgumentException.class, () -> orderService.placeOrder(order));
    }

    // ── @CsvSource ────────────────────────────────────────────────────────
    @ParameterizedTest(name = "status={0} → refundable={1}")
    @CsvSource({
        "PENDING,   true",
        "CONFIRMED, true",
        "SHIPPED,   false",
        "DELIVERED, false"
    })
    void isRefundable(String status, boolean expected) {
        assertEquals(expected, orderService.isRefundable(status));
    }

    // ── @MethodSource ─────────────────────────────────────────────────────
    @ParameterizedTest
    @MethodSource("largeOrderProvider")
    void placeOrder_appliesDiscountForLargeOrders(BigDecimal amount, BigDecimal expectedCharge) {
        var order = new Order("cust-vip", amount);
        when(repository.save(any())).thenAnswer(inv -> ((Order) inv.getArgument(0)).withId(99L));
        when(paymentGateway.charge(eq("cust-vip"), eq(expectedCharge))).thenReturn("txn-disc");

        Receipt receipt = orderService.placeOrder(order);

        // Capture the actual amount charged for a detailed failure message
        ArgumentCaptor<BigDecimal> captor = ArgumentCaptor.forClass(BigDecimal.class);
        verify(paymentGateway).charge(eq("cust-vip"), captor.capture());
        assertEquals(expectedCharge, captor.getValue(),
            "Expected discounted charge for amount=" + amount);
    }

    static Stream<Arguments> largeOrderProvider() {
        return Stream.of(
            Arguments.of(new BigDecimal("500.00"), new BigDecimal("475.00")),  // 5 % off
            Arguments.of(new BigDecimal("1000.00"), new BigDecimal("900.00")) // 10 % off
        );
    }
}
