package com.example.hibernate.entity;

import jakarta.persistence.*;
import java.util.ArrayList;
import java.util.List;

@Entity
@Table(name = "promotions")
public class Promotion {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String code;

    @Column(nullable = false)
    private int discountPercent;

    @ManyToMany(mappedBy = "promotions")
    private List<Order> orders = new ArrayList<>();

    protected Promotion() {}

    public Promotion(String code, int discountPercent) {
        this.code = code;
        this.discountPercent = discountPercent;
    }

    public Long getId() { return id; }
    public String getCode() { return code; }
    public int getDiscountPercent() { return discountPercent; }
}
