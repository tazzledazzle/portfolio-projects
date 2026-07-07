package com.example.hibernate;

import com.example.hibernate.entity.*;
import com.example.hibernate.repository.OrderRepository;
import jakarta.persistence.*;
import java.math.BigDecimal;

public class Main {

    public static void main(String[] args) {
        EntityManagerFactory emf = Persistence.createEntityManagerFactory("example-pu");
        EntityManager em = emf.createEntityManager();

        try {
            EntityTransaction tx = em.getTransaction();

            // ── persist ───────────────────────────────────────────────────
            tx.begin();

            var address  = new Address("123 Main St", "Springfield", "US");
            var customer = new Customer("Alice", "alice@example.com");
            customer.setAddress(address);

            var order = new Order(customer, "PENDING");
            order.addLineItem(new LineItem("Widget A", 2, new BigDecimal("9.99")));
            order.addLineItem(new LineItem("Widget B", 1, new BigDecimal("24.99")));

            var promo = new Promotion("SUMMER10", 10);
            order.getPromotions().add(promo);

            em.persist(customer);   // cascades to address, order, line items
            em.persist(promo);

            tx.commit();

            Long orderId = order.getId();

            // ── detach + merge ────────────────────────────────────────────
            em.detach(order);
            order.setStatus("CONFIRMED");

            tx.begin();
            Order managed = em.merge(order);   // reattach and UPDATE
            tx.commit();

            // ── refresh ───────────────────────────────────────────────────
            em.refresh(managed);   // discard any in-memory drift, reload from DB

            // ── JPQL query ────────────────────────────────────────────────
            var repo = new OrderRepository(em);
            repo.findByCustomer(customer)
                .forEach(o -> System.out.println("Order #" + o.getId() + " — " + o.getStatus()));

            // ── Criteria API query ─────────────────────────────────────────
            repo.findByStatus("CONFIRMED")
                .forEach(o -> System.out.println("Confirmed order: " + o.getId()));

            // ── remove ────────────────────────────────────────────────────
            tx.begin();
            repo.delete(managed);   // cascades removal to line items (orphanRemoval)
            tx.commit();

        } finally {
            em.close();
            emf.close();
        }
    }
}
