# FastAPI Online Bookstore

## **1. Requirements**

### **Functional**

1. **Catalog Browsing**: List books by category, author, bestseller, new releases.
2. **Search**: Full‐text search over title, author, description; faceted by genre, price, rating.
3. **Book Details**: View metadata (title, author, ISBN, description, reviews), availability, price.
4. **Shopping Cart**: Add/remove books; view cart contents.
5. **Checkout & Payment**: Capture shipping info, payment (credit card, wallet), apply promotions/discounts, generate order.
6. **Order Management**: View order history, track shipment status.
7. **Inventory Updates**: Decrement stock on successful purchase; prevent oversell.

### **Non-functional**

- **Latency**: p95 catalog/listing & search ≤100 ms; checkout end-to-end ≤2 s.
- **Throughput**: ≥2 K browse/sec, ≥1 K search/sec, ≥500 checkouts/sec.
- **Availability**: ≥99.9% overall, ≥99.99% for read paths.
- **Scalability**: Handle seasonal spikes (e.g. holidays).
- **Security & Compliance**: PCI DSS for payments; encrypt PII in transit & at rest.

---

## **2. High-Level Architecture**

```bash
      ┌──────────┐        ┌──────────────┐        ┌───────────┐
      │ Clients  │─HTTPS─▶│ API Gateway  │▶──┐    │ Identity  │
      └──────────┘        └─────┬────────┘   │    │ Service   │
                                   │         │    └───────────┘
           ┌───────────────────────┼─────────┼────────────────────────┐
           │                       ▼         ▼                        │
┌──────────▼─────────┐   ┌─────────▼─────────┐   ┌─────────────────┐  │
│ Catalog Service    │   │ Search Service    │   │ Cart & Order    │  │
│ (Books, Metadata)  │   │ (Elasticsearch)   │   │ Service         │  │
└──────────┬─────────┘   └─────────┬─────────┘   └───┬─────────────┘  │
           │                       │               │                │
           ▼                       ▼               ▼                │
   ┌───────────────┐        ┌──────────────┐ ┌─────────────┐        │
   │ Metadata DB   │        │ Search Index │ │ Orders DB   │        │
   └───────────────┘        └──────────────┘ └─────────────┘        │
           ▲                                               ▲        │
           │                                               │        │
┌──────────┴─────────┐                             ┌───────┴────────┐ │
│ Inventory Service  │                             │ Payment &     │ │
│ (Stock Levels)     │◀─────┐                      │ Shipping Svc  │ │
└────────────────────┘      │                      └───────────────┘ │
                            │                                         │
                           ┌┴┐                                        │
                           │Cache (Redis)                             │
                           └─┘                                        │
           (optional)                                               ──┘
```

---

## **3. Core Components**

- **API Gateway**: TLS termination, routing, rate-limit, auth token check.
- **Catalog Service**: CRUD on book metadata; reads from Metadata DB; writes update Search Index and Cache.
- **Search Service**: Query Elasticsearch; returns ranked results with facets.
- **Inventory Service**: Manages stock counts; atomic decrement; publishes low-stock alerts.
- **Cart & Order Service**:
  - **Cart**: ephemeral per-user store (Redis).
  - **Order**: creates orders in Orders DB; orchestrates inventory reservation, payment capture, shipping request.
- **Payment & Shipping Service**: integrates with external PSPs (Stripe/PayPal) and carriers (FedEx, UPS).
- **Identity Service**: user registration, login, JWT issuance.

---

## **4. Data Models**

> Currently representing database storage as a CSV

| **Books** |  
| --- | --- | --- | --- |
| book_id UUID (PK) |  
| title STRING |  
| authors [STRING] |  
| price_cents INT |  
| categories [STR] |  
| description TEXT |  
| isbn STRING |  
| rating FLOAT |  

| **Orders** |  |
| order_id UUID (PK) |  |
| user_id UUID (FK) |  |
| order_date TIMESTAMP |  |
| status ENUM | {PENDING, PAID…} |
| total_cents INT |  |
| … |  |
|  |  |

| **Order_Items** |  |
  | order_id FK |  |
  | book_id FK |  |
 | quantity INT |  |
| **Inventory** |
| book_id PK (FK) |
| stock INT |

**Search Index** documents mirror book metadata plus popularity signals.

---

## **5. API Endpoints**

```bash
GET    /v1/books                 → paginated catalog (filters: category, author)
GET    /v1/books/{book_id}       → book details + availability
GET    /v1/search?q=…&facets=…   → free-text + faceting
POST   /v1/cart                  → { book_id, qty }
GET    /v1/cart                  → current cart
DELETE /v1/cart/{book_id}        → remove item
POST   /v1/checkout              → { payment_info, shipping_addr }
GET    /v1/orders/{order_id}     → order status/history
```

---

## **6. Read & Write Flows**

- **Browsing/Search**
    1. Client → API GW → Cache lookup for /books or /search.
    2. On cache miss → Catalog Service / Search Service → Cache result.
    3. Return list with pagination & facets.
- **Checkout**
    1. Client sends POST /checkout → Cart & Order Service.
    2. **Transaction**:
        - **Reserve Inventory**: call Inventory Service to decrement stock; if insufficient, abort.
        - **Process Payment**: synchronous call to Payment Service; on failure, roll back inventory.
        - **Persist Order** in Orders DB; emit “order_created” event.
        - **Notify Shipping**: event-driven to Shipping Service.
    3. Return order confirmation to client; send email.

---

## **7. Scalability & Partitioning**

- **Stateless Services** behind auto-scaling groups (Kubernetes/ECS).
- **Metadata DB**: shard by book_id or category; read replicas for catalog reads.
- **Search**: Elasticsearch cluster with shards/replicas; autoscale indexing nodes.
- **Cache**: Redis cluster for hot catalog pages, search hotspots, and carts.
- **Inventory**: strongly consistent store (e.g. Spanner, CockroachDB) or use per-book partitioned counters.

---

## **8. Consistency & Trade-Offs**

| **Concern** | **Strict Consistency** | **Eventual Consistency** |
| --- | --- | --- |
| Inventory updates | Two-phase commit on checkout | Asynchronous reservation + compensating rollback |
| Catalog freshness | Immediate reflecting of edits | Indexed/Cache refresh lag (~sec) |

- Use **synchronous inventory** decrement to prevent oversell, accepting slight checkout latency.
- Use **async catalog updates** to Elasticsearch & cache to maximize read throughput.

---

## **9. Monitoring & Alerting**

- **Metrics**:
  - QPS & latency per endpoint (p50/p95/p99).
  - Cache hit ratios for catalog & search.
  - Checkout success vs. failure rates.
  - Inventory reservation failures.
- **Logs & Traces**:
  - End-to-end tracing for checkout flow.
  - Error logs (DB timeouts, payment declines).
- **Alerts**:
  - Checkout error rate >1%.
  - Inventory service error spike.
  - Search cluster index backlog.

---

## **10. Security & Compliance**

- **PCI DSS** scope: tokenize credit cards; no raw card data in our systems.
- **Authentication & Authorization**: JWT with short TTL; refresh tokens.
- **Input Validation & Sanitization**: prevent injection on search/catalog.
- **HTTPS Everywhere**; HSTS.
- **PII Encryption** at rest for user/shipping data.
