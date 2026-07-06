package com.example.jooq;

import org.jooq.*;
import org.jooq.impl.DSL;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

import static com.example.jooq.generated.Tables.*;
import static org.jooq.impl.DSL.*;

// Generated DSL classes (Tables.ORDERS, Tables.LINE_ITEMS, etc.) come from jOOQ's
// code-generator run against the live schema — they encode column types at compile time.
public class OrderRepository {

    private final DSLContext dsl;

    public OrderRepository(DSLContext dsl) {
        this.dsl = dsl;
    }

    // ── SELECT ─────────────────────────────────────────────────────────────
    // Result<Record> is a list of Row-like Records; Field<T> carries the column type.
    public Result<Record3<Long, String, LocalDateTime>> findOrderSummaries(String status) {
        return dsl
            .select(ORDERS.ID, ORDERS.STATUS, ORDERS.CREATED_AT)
            .from(ORDERS)
            .where(ORDERS.STATUS.eq(status))
            .orderBy(ORDERS.CREATED_AT.desc())
            .fetch();
    }

    // JOIN with aggregate — returns a typed Record for each row
    public Result<Record3<Long, String, BigDecimal>> orderTotals() {
        return dsl
            .select(
                ORDERS.ID,
                ORDERS.STATUS,
                sum(LINE_ITEMS.UNIT_PRICE.mul(LINE_ITEMS.QUANTITY)).as("total")
            )
            .from(ORDERS)
            .join(LINE_ITEMS).on(LINE_ITEMS.ORDER_ID.eq(ORDERS.ID))
            .groupBy(ORDERS.ID, ORDERS.STATUS)
            .fetch();
    }

    // Fetch a single strongly-typed Record
    public OrdersRecord findById(long id) {
        return dsl
            .selectFrom(ORDERS)
            .where(ORDERS.ID.eq(id))
            .fetchOne();
    }

    // ── INSERT ─────────────────────────────────────────────────────────────
    public long insertOrder(long customerId, String status) {
        return dsl
            .insertInto(ORDERS, ORDERS.CUSTOMER_ID, ORDERS.STATUS, ORDERS.CREATED_AT)
            .values(customerId, status, LocalDateTime.now())
            .returningResult(ORDERS.ID)
            .fetchOne()
            .value1();
    }

    // ── UPDATE ─────────────────────────────────────────────────────────────
    public int updateStatus(long orderId, String newStatus) {
        return dsl
            .update(ORDERS)
            .set(ORDERS.STATUS, newStatus)
            .where(ORDERS.ID.eq(orderId))
            .execute();
    }

    // ── DELETE ─────────────────────────────────────────────────────────────
    public int deleteOrder(long orderId) {
        return dsl
            .deleteFrom(ORDERS)
            .where(ORDERS.ID.eq(orderId))
            .execute();
    }

    // ── TRANSACTION ────────────────────────────────────────────────────────
    // DSLContext.transaction() wraps the lambda in a single DB transaction;
    // any unchecked exception triggers automatic rollback.
    public void transferLineItem(long fromOrderId, long toOrderId, long lineItemId) {
        dsl.transaction(cfg -> {
            DSLContext tx = DSL.using(cfg);

            int updated = tx
                .update(LINE_ITEMS)
                .set(LINE_ITEMS.ORDER_ID, toOrderId)
                .where(LINE_ITEMS.ID.eq(lineItemId)
                    .and(LINE_ITEMS.ORDER_ID.eq(fromOrderId)))
                .execute();

            if (updated == 0) {
                throw new IllegalArgumentException(
                    "Line item " + lineItemId + " not found on order " + fromOrderId);
            }

            // Recalculate totals for both orders (simplified: just log)
            tx.execute("CALL recalculate_order_total({0})", fromOrderId);
            tx.execute("CALL recalculate_order_total({0})", toOrderId);
        });
    }
}
