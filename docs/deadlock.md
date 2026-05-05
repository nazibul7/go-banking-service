# Deadlock — how it actually happens in concurrent money transfers

While testing my Go + PostgreSQL banking API, I hit a real production-style deadlock issue during concurrent transfers.

## Where it happened

In the `/account/transfer` endpoint.

## How I reproduced it

Terminal 1:

```bash
while true; do
curl -X POST http://localhost:8080/account/transfer \
-H "Content-Type: application/json" \
-d '{"from_id":1,"to_id":2,"amount":1}'
done
```

Terminal 2:

```bash
while true; do
curl -X POST http://localhost:8080/account/transfer \
-H "Content-Type: application/json" \
-d '{"from_id":2,"to_id":1,"amount":1}'
done
```

After running both concurrently:

```text
pq: deadlock detected
```

started appearing continuously.

---

# What was actually happening internally

Two opposite transfers were executing at the same time:

```text
T1: transfer 1 → 2
T2: transfer 2 → 1
```

## Step 1

```text
T1 locks account 1 ✓
T2 locks account 2 ✓
```

## Step 2

```text
T1 tries to lock account 2 → WAITING
T2 tries to lock account 1 → WAITING
```

Now both transactions are waiting for each other.

This creates a circular dependency:

```text
T1 → waiting for T2
T2 → waiting for T1
```

PostgreSQL detects the cycle and aborts one transaction:

```text
pq: deadlock detected
```

---

# The fix

I fixed it using deterministic locking:

```go
if fromID < toID {
    // lock smaller ID first
} else {
    // still lock smaller ID first
}
```

Now both transactions acquire locks in the same order:

```text
T1 (1→2) → lock 1 then 2
T2 (2→1) → lock 1 then 2
```

No circular wait.

No deadlock.

Only normal lock waiting.

---

# This technique is called

* Deterministic locking
* Consistent lock ordering
* Ordered locking
* Canonical lock order

This pattern is heavily used in:

* banking systems
* payment systems
* inventory systems
* distributed systems
* operating systems

---

# Biggest learning

Concurrency bugs are not always race conditions.

Sometimes:

* transactions are correct
* queries are atomic
* locks are working properly

…but lock acquisition order itself becomes the problem.

Understanding database locking behavior changed how I think about backend systems.
