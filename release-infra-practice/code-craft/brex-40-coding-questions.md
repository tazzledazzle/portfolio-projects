# 40 Brex Coding Questions — 20 Practical + 20 Algorithmic

## What your recruiter's FAQ actually tells you (and where it agrees/disagrees with field reports)

Your FAQ screenshots say the coding round is **"a practical Data Structure/Algorithm or Data Modeling question"**
that will **not** try to trick you the way a traditional LeetCode question does, graded on: clean code in your chosen language,
reaching a *running* solution (not necessarily the optimal one on the first pass), a deterministic debugging strategy,
and clear communication of your assumptions.

That framing mostly matches what real candidates report — but not entirely. A meaningful minority of accounts describe
rounds that were flatly algorithmic and not gentle: a phone screen scored as **"Leetcode Hard Level"** (interviewer gave hints),
a Software Engineer II account that got a **"DP hard problem from leetcode"**, and a 1Point3Acres log titled simply
**"Calculator Problem."** Treat your official FAQ as the *intended* design of the round, and treat the harder algorithmic
reports below as the real variance you should still be ready for — the two lists aren't in conflict so much as covering
the intended center of the distribution vs. its observed tail.

Both lists below are sourced, with a confidence tag per item:
- **[Named]** — a specific, named/described problem tied to an individual account or aggregator entry (highest confidence)
- **[Pattern]** — constructed to match a repeatedly-confirmed pattern (base problem + 2-3 extension layers, fintech domain)
  where no single named example was found, so the shape is real even if this exact wording isn't a confirmed verbatim question

---

## List 1: Practical / "Build Something" (20)

Matches Brex's own internal name for this format: the **"Craft"** round — start with a base problem, extend it with 2-3
additional feature layers, graded on clean code, data modeling, and edge-case handling rather than algorithmic optimality.

| #  | Question                                                                                                                                                                                           | Confidence | Source                                                                                                                           |
|----|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------|----------------------------------------------------------------------------------------------------------------------------------|
| 1  | **Can Purchase Card** — implement a function deciding whether a card purchase is allowed, then extend with idempotency handling and reconciling against a transaction audit log                    | [Named]    | TechPrep                                                                                                                         |
| 2  | **Moving Average of Transaction Volumes** — compute a rolling average over a stream of transaction volume data                                                                                     | [Named]    | TechPrep                                                                                                                         |
| 3  | **Gem Card Purchases with Color Discounts** — implement purchase logic for a gem-based card system with five gem colors, each with its own discount rule                                           | [Named]    | TechPrep / PracHub                                                                                                               |
| 4  | **Credit-limit checker** — build a small back-end service from scratch that checks and enforces a credit limit                                                                                     | [Named]    | TechPrep                                                                                                                         |
| 5  | **Transaction parser** — build a small back-end service from scratch that parses incoming transaction records                                                                                      | [Named]    | TechPrep                                                                                                                         |
| 6  | **Points-from-Transactions Service** — design/implement a loyalty-points service (1 USD = 1.5 points) supporting awarding, balance queries, and redemption with double-spend prevention            | [Named]    | PracHub                                                                                                                          |
| 7  | **AI-Assisted Full-Stack Transactions App** — build a minimal full-stack web app supporting basic monetary transactions                                                                            | [Named]    | PracHub                                                                                                                          |
| 8  | **React color multi-select dropdown** — implement a UI letting users multi-select colors from a dropdown and view details, extended with filtering/state management                                | [Named]    | PracHub                                                                                                                          |
| 9  | **Debug and Explain Fixes** — given a small codebase with failing unit tests, find root causes, fix the bugs, and explain each defect found                                                        | [Named]    | PracHub / TechPrep / multiple Glassdoor accounts                                                                                 |
| 10 | **Idempotent card-transaction processor** — reject duplicate submissions by idempotency key, return 429 if a card exceeds 5 transactions/minute, log every accepted transaction to a simple ledger | [Named]    | JobsByCulture                                                                                                                    |
| 11 | **REST + JSON + sort service** — make REST calls, parse the JSON responses, correctly convert/normalize date-time fields across sources, then sort the results                                     | [Named]    | Blind + Glassdoor (2 independent accounts)                                                                                       |
| 12 | **Word-wrap / line justification** — split a list of words into lines of exactly N characters, keeping whole words intact and distributing extra spaces across the line                            | [Named]    | Glassdoor (2019 Frontend review)                                                                                                 |
| 13 | **Calculator** — implement a working calculator (exact scope undisclosed in the source log, but confirmed as an actual phone-screen topic)                                                         | [Named]    | 1Point3Acres                                                                                                                     |
| 14 | **"Vibe code" a P2P payment app** — open-ended AI-assisted build task: construct a peer-to-peer payment web app resembling PayPal, scope left entirely to the candidate                            | [Named]    | Glassdoor (Jan 2026 AI round)                                                                                                    |
| 15 | **Transaction Processing Engine** — implement the core of a transaction-processing engine with attention to exactly-once handling                                                                  | [Named]    | TechPrep (also appears as a system-design topic)                                                                                 |
| 16 | **Expense auto-approval flow** — categorize an incoming expense, then extend with a rule that auto-approves under a threshold and routes above it to manager approval                              | [Pattern]  | Constructed — matches Brex's own expense-management product domain                                                               |
| 17 | **Card freeze/unfreeze state machine** — implement freeze/unfreeze transitions for a card with a full audit trail of state changes                                                                 | [Pattern]  | Constructed — matches TechPrep's explicitly named "state machines" category                                                      |
| 18 | **Rate-limited webhook delivery** — deliver webhook events to a subscriber with retry/backoff, extended to handle a subscriber that's intermittently down                                          | [Pattern]  | Constructed — matches the confirmed "practical API" pattern                                                                      |
| 19 | **Multi-currency conversion with correct rounding** — convert an expense amount between currencies without floating-point drift, extended with a batch-conversion endpoint                         | [Pattern]  | Constructed — matches the "monetary precision and rounding" theme explicitly called out in Brex-vs-Mercury interview comparisons |
| 20 | **Recurring/subscription payment scheduler** — process a recurring payment schedule, extended to handle a missed billing cycle and retry logic                                                     | [Pattern]  | Constructed — matches TechPrep's explicitly named "scheduling" category                                                          |

---

## List 2: Algorithmic / LeetCode-Style (20)

Sourced as a single set from an aggregator that tracks Brex-tagged
LeetCode-style questions specifically (InterviewSolver) — this is the
one part of this doc that's **[Aggregator]**-tier rather than tied to
individual first-hand accounts, so treat frequencies as directional. The
stated distribution: 4 Easy / 14 Medium / 2 Hard, skewed toward **Array,
Hash Table, Binary Search, Sorting, and String** — which lines up with
what your FAQ calls "practical Data Structure/Algorithm," not contest-style
tricks.

| #  | Problem                                                | Difficulty | Primary topic                    |
|----|--------------------------------------------------------|------------|----------------------------------|
| 1  | House Robber II                                        | Medium     | Dynamic Programming              |
| 2  | K-diff Pairs in an Array                               | Medium     | Array / Hash Table               |
| 3  | Majority Element                                       | Easy       | Array / Hash Table               |
| 4  | Check If N and Its Double Exist                        | Easy       | Array / Hash Table               |
| 5  | Search a 2D Matrix                                     | Medium     | Binary Search                    |
| 6  | Minimum Operations to Make Character Frequencies Equal | Hard       | String / Hash Table              |
| 7  | Sqrt(x)                                                | Easy       | Binary Search / Math             |
| 8  | Reverse Words in a String                              | Medium     | String                           |
| 9  | Plus One                                               | Easy       | Array                            |
| 10 | Decode String                                          | Medium     | Stack / String                   |
| 11 | Shortest Word Distance II                              | Medium     | Hash Table / Design              |
| 12 | Frequency of the Most Frequent Element                 | Medium     | Array / Sorting / Sliding Window |
| 13 | Unique Paths                                           | Medium     | Dynamic Programming              |
| 14 | Zigzag Conversion                                      | Medium     | String                           |
| 15 | Count Unhappy Friends                                  | Medium     | Array / Simulation               |
| 16 | Kth Smallest Element in a Sorted Matrix                | Medium     | Binary Search / Heap             |
| 17 | Rotting Oranges                                        | Medium     | BFS / Matrix                     |
| 18 | Line Reflection                                        | Medium     | Hash Table / Geometry            |
| 19 | Vertical Order Traversal of a Binary Tree              | Hard       | Tree / BFS                       |
| 20 | Maximum Value at a Given Index in a Bounded Array      | Medium     | Binary Search / Greedy           |

**Corroborating first-hand signal**
(not part of the counted 20, but confirms this general shape is real): a "DP hard problem from leetcode"in one Software
Engineer II phone screen, and a separate "Leetcode Hard Level" phone screen — both consistent with the Hard-tier
DP/array-family problems on the list above, even though neither account named the exact problem.

---

## Reconciling this with the FAQ's grading criteria

Whichever list a given round pulls from, your FAQ tells you exactly what's being scored — worth re-reading the moment
you're handed either type of problem:

1. **Get to something that runs.** The FAQ explicitly says you don't need the optimal solution on the first pass — for
    List 2 especially, ship a correct brute force before optimizing, and narrate the upgrade path out loud rather than
    jumping straight to the optimal version silently.
2. **Show a deterministic debugging strategy.** This is called out explicitly as graded — for List 1's build tasks,
    this means when something breaks, narrate a systematic narrowing process (check inputs, isolate the failing layer,
    form a hypothesis) rather than randomly poking at the code.
3. **Talk before you type.** The FAQ's warning against "jumping straight into coding" matches everything in the field
    reports about interviewers wanting stated assumptions up front — for List 1's fintech-flavored problems specifically,
    stating your assumptions about idempotency, rounding, and consistency *before* writing code is itself part of the
    graded signal, not just throat-clearing.
4. **Data modeling is graded as its own thing**, distinct from algorithm correctness — worth explicitly sketching your
    data model (what fields, what invariants, what's mutable) for any List 1 problem before writing the first line of logic.
