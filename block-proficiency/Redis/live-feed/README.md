# Live Notification Feed — Kotlin + Redis

A production-pattern notification system built on Redis **Pub/Sub** and **List** data structures,
using **Lettuce** as the JVM client with Kotlin coroutine-friendly async support.

---

## What this project teaches

| Redis concept | Where it's used |
|---|---|
| **Pub/Sub** (`PUBLISH`, `SUBSCRIBE`) | Fan-out: one publisher → all active subscribers |
| **List** (`LPUSH`, `LRANGE`, `LTRIM`) | Per-user inbox: durable, ordered, pageable |
| **`EXPIRE`** | Auto-expire inactive inboxes after 7 days |
| **`TTL`** | Inspect remaining inbox lifetime |
| **`DEL`** | Clear an inbox on demand |
| **Key namespacing** | `inbox:{userId}` — one List key per user |

---

## Architecture

```
┌────────────────┐     PUBLISH notifications:events     ┌──────────────────────────┐
│ NotificationPub│ ──────────────────────────────────►  │ Redis Pub/Sub channel    │
│ lisher         │                                       │ "notifications:events"   │
└────────────────┘                                       └──────────┬───────────────┘
                                                                    │ message delivery
                                                                    ▼
                                                         ┌──────────────────────────┐
                                                         │ NotificationSubscriber   │
                                                         │ (RedisPubSubAdapter)     │
                                                         └──────────┬───────────────┘
                                                                    │ inboxService.push()
                                                                    ▼
┌────────────────┐    LRANGE inbox:{userId} 0 19        ┌──────────────────────────┐
│ API / Demo     │ ◄──────────────────────────────────  │ InboxService             │
│ (reader)       │                                       │                          │
└────────────────┘                                       │ LPUSH  inbox:{userId}    │
                                                         │ LTRIM  inbox:{userId} …  │
                                                         │ EXPIRE inbox:{userId} …  │
                                                         └──────────────────────────┘
```

### Why three Redis connections?

```
publisherConn  (StatefulRedisPubSubConnection)  — PUBLISH only
subscriberConn (StatefulRedisPubSubConnection)  — SUBSCRIBE only
commandConn    (StatefulRedisConnection)         — LPUSH / LRANGE / LTRIM / EXPIRE / TTL
```

Redis **requires** that a connection in SUBSCRIBE mode can only issue Pub/Sub commands.
Mixing PUBLISH or LPUSH on a subscribed connection causes protocol errors.
Lettuce connections are thread-safe, so one `commandConn` is shared across the whole app.

---

## Project structure

```
notification-feed/
├── build.gradle.kts
├── settings.gradle.kts
└── src/
    ├── main/
    │   ├── kotlin/com/skidroad/notifeed/
    │   │   ├── Main.kt                          ← entry point + demo
    │   │   ├── RedisConfig.kt                   ← Lettuce connection factory
    │   │   ├── model/
    │   │   │   ├── Notification.kt              ← domain model
    │   │   │   └── Serialisation.kt             ← Jackson JSON helpers
    │   │   ├── publisher/
    │   │   │   └── NotificationPublisher.kt     ← wraps PUBLISH
    │   │   ├── subscriber/
    │   │   │   └── NotificationSubscriber.kt    ← RedisPubSubAdapter
    │   │   ├── inbox/
    │   │   │   └── InboxService.kt              ← LPUSH / LRANGE / LTRIM / EXPIRE
    │   │   └── api/
    │   │       └── NotificationService.kt       ← high-level facade
    │   └── resources/
    │       └── logback.xml
    └── test/
        └── kotlin/com/skidroad/notifeed/
            ├── InboxServiceTest.kt
            └── NotificationPublisherTest.kt
```

---

## Step-by-step setup

### Step 1 — Prerequisites

- **JDK 17+** (`java -version`)
- **Gradle 8+** (or use the wrapper: `./gradlew`)
- **Redis 6+** running locally on port 6379

```bash
# macOS
brew install redis
brew services start redis

# Ubuntu / Debian
sudo apt install redis-server
sudo systemctl start redis

# Docker (quickest)
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

Verify Redis is up:
```bash
redis-cli ping
# → PONG
```

### Step 2 — Clone / open the project

```bash
cd notification-feed
```

### Step 3 — Run the demo

```bash
./gradlew run
```

Expected output (timestamps omitted):
```
=== Live Notification Feed — startup ===
Opening Redis command connection to localhost:6379
Opening Redis Pub/Sub connection to localhost:6379
Opening Redis Pub/Sub connection to localhost:6379
Subscriber started — listening on channel='notifications:events'
SUBSCRIBE confirmed  channel='notifications:events'  activeSubscriptions=1

--- Publishing notifications for two users ---
PUBLISH → channel='notifications:events'  type=ORDER_SHIPPED   userId=alice  receivers=1
→ Routing notification  type=ORDER_SHIPPED   userId=alice
PUBLISH → channel='notifications:events'  type=MENTION         userId=alice  receivers=1
→ Routing notification  type=MENTION         userId=alice
PUBLISH → channel='notifications:events'  type=FRIEND_REQUEST  userId=bob    receivers=1
→ Routing notification  type=FRIEND_REQUEST  userId=bob
PUBLISH → channel='notifications:events'  type=SYSTEM_ALERT    userId=bob    receivers=1
→ Routing notification  type=SYSTEM_ALERT    userId=bob

--- Reading inboxes ---
  Inbox for 'alice'  (size=2  ttl=604799s)
    [0] MENTION               | Bob mentioned you in a comment
    [1] ORDER_SHIPPED         | Your order #1042 has shipped!

  Inbox for 'bob'    (size=2  ttl=604800s)
    [0] SYSTEM_ALERT          | Scheduled maintenance tonight at 11 PM UTC
    [1] FRIEND_REQUEST        | Alice sent you a friend request

--- Pagination demo ---
Alice inbox size: 27
Page 0 (newest 20): ...
Page 1 (next 7): ...

--- TTL inspection ---
alice inbox TTL: 604799 seconds (≈7 days)
bob   inbox TTL: 604800 seconds

--- Clearing bob's inbox ---
bob inbox size after clear: 0
bob inbox TTL  after clear: -2   (−2 = key gone)

Total Pub/Sub messages received by subscriber: 29
```

### Step 4 — Run the tests

```bash
./gradlew test
```

Tests use MockK — no real Redis needed.

```
InboxServiceTest
  ✓ push calls LPUSH, LTRIM, EXPIRE and LLEN in order
  ✓ push uses correct inbox key pattern
  ✓ getPage returns correct page from LRANGE
  ✓ getPage computes correct LRANGE start and stop for page 2
  ✓ getPage skips malformed JSON and returns valid items
  ✓ size delegates to LLEN
  ✓ ttl delegates to TTL command
  ✓ clear calls DEL on the inbox key

NotificationPublisherTest
  ✓ publish calls PUBLISH on the correct channel
  ✓ publish returns 0 when no subscribers are active
  ✓ publish serialises the notification as JSON containing the userId
```

### Step 5 — Inspect Redis live (optional)

While the demo runs (or after), open `redis-cli` to observe the state:

```bash
redis-cli

# List all inbox keys
KEYS inbox:*
# → 1) "inbox:alice"
# → 2) "inbox:bob"

# Inspect Alice's inbox (newest first)
LRANGE inbox:alice 0 -1

# Check length
LLEN inbox:alice

# Check TTL
TTL inbox:alice
# → 604799 (seconds remaining)

# Monitor all Redis commands in real time (run BEFORE ./gradlew run)
MONITOR
```

---

## Key Redis commands reference

```
PUBLISH channel message        Fan-out message to all subscribers
SUBSCRIBE channel              Enter subscription mode; receive messages via listener

LPUSH key value                Prepend value to list (newest → head)
LRANGE key start stop          Read a range of list elements (0-indexed)
LTRIM key start stop           Keep only elements in range; drop the rest
LLEN key                       Number of elements in the list

EXPIRE key seconds             Set a TTL in seconds
TTL key                        Remaining TTL (-2 = gone, -1 = no expiry)
DEL key                        Delete the key
```

---

## Extending this project

**Add read/unread state**
Store a `read` boolean per notification. Since notifications are JSON in a List,
you'd need to scan + replace a specific element. For heavy read-tracking, consider a
parallel Set of read notification IDs (`SADD read:{userId} {notifId}`).

**Durable persistence**
Pub/Sub is fire-and-forget: if the subscriber is down when a message arrives, it's lost.
For durability, consider Redis Streams (`XADD` / `XREADGROUP`) — they provide consumer
groups with acknowledgement semantics.

**Multiple subscribers / fan-out per user**
The current system uses one global subscriber. In a scaled deployment, each app instance
runs its own subscriber. Pub/Sub already fan-outs to all of them — each writes to the
same `inbox:{userId}` key. LPUSH on a List is atomic, so no race conditions.

**Reactive / coroutine API**
Lettuce's `pubSubConn.reactive()` returns an RxJava Publisher; bridge to Kotlin Flow
via `kotlinx-coroutines-reactive`'s `asFlow()` for non-blocking processing pipelines.