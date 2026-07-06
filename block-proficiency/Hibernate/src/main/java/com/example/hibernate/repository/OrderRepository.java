package com.example.hibernate.repository;

import com.example.hibernate.entity.Customer;
import com.example.hibernate.entity.Order;
import jakarta.persistence.*;
import jakarta.persistence.criteria.*;
import java.util.List;
import java.util.Optional;

public class OrderRepository {

    private final EntityManager em;

    public OrderRepository(EntityManager em) {
        this.em = em;
    }

    // persist — make a transient entity managed and schedule INSERT
    public void save(Order order) {
        em.persist(order);
    }

    // merge — reattach a detached entity; returns the managed copy
    public Order update(Order detachedOrder) {
        return em.merge(detachedOrder);
    }

    // remove — schedule DELETE; entity must be managed
    public void delete(Order order) {
        Order managed = em.contains(order) ? order : em.merge(order);
        em.remove(managed);
    }

    // refresh — discard in-memory changes and reload from DB
    public void refresh(Order order) {
        em.refresh(order);
    }

    // detach — remove from persistence context without deleting from DB
    public void detach(Order order) {
        em.detach(order);
    }

    // JPQL query — finds all orders for a given customer, eagerly fetching line items
    public List<Order> findByCustomer(Customer customer) {
        return em.createQuery(
            "SELECT o FROM Order o " +
            "JOIN FETCH o.lineItems " +
            "WHERE o.customer = :customer " +
            "ORDER BY o.createdAt DESC",
            Order.class
        )
        .setParameter("customer", customer)
        .getResultList();
    }

    // JPQL — aggregate: total revenue per status
    public List<Object[]> totalRevenueByStatus() {
        return em.createQuery(
            "SELECT o.status, SUM(li.unitPrice * li.quantity) " +
            "FROM Order o JOIN o.lineItems li " +
            "GROUP BY o.status",
            Object[].class
        ).getResultList();
    }

    // Criteria API — type-safe dynamic query; filters by status when provided
    public List<Order> findByStatus(String status) {
        CriteriaBuilder cb = em.getCriteriaBuilder();
        CriteriaQuery<Order> cq = cb.createQuery(Order.class);
        Root<Order> root = cq.from(Order.class);

        if (status != null && !status.isBlank()) {
            cq.where(cb.equal(root.get("status"), status));
        }

        cq.orderBy(cb.desc(root.get("createdAt")));
        return em.createQuery(cq).getResultList();
    }

    public Optional<Order> findById(Long id) {
        return Optional.ofNullable(em.find(Order.class, id));
    }
}
