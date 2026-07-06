package com.example.hibernate.entity;

import jakarta.persistence.*;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Entity
@Table(name = "orders")
public class Order {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;

    @Column(nullable = false)
    private String status;

    // Many orders belong to one customer; FK lives on this side
    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "customer_id")
    private Customer customer;

    // One order has many line items; orphanRemoval keeps the DB clean
    @OneToMany(mappedBy = "order", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<LineItem> lineItems = new ArrayList<>();

    @ManyToMany(cascade = {CascadeType.PERSIST, CascadeType.MERGE})
    @JoinTable(
        name = "order_promotions",
        joinColumns = @JoinColumn(name = "order_id"),
        inverseJoinColumns = @JoinColumn(name = "promotion_id")
    )
    private List<Promotion> promotions = new ArrayList<>();

    protected Order() {}

    public Order(Customer customer, String status) {
        this.customer = customer;
        this.status = status;
        this.createdAt = LocalDateTime.now();
    }

    public void addLineItem(LineItem item) {
        item.setOrder(this);
        lineItems.add(item);
    }

    // getters / setters omitted for brevity
    public Long getId() { return id; }
    public Customer getCustomer() { return customer; }
    public List<LineItem> getLineItems() { return lineItems; }
    public List<Promotion> getPromotions() { return promotions; }
    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }
}
