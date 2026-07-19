# Concurrency Bug Investigation: Idempotency in Go Banking API

## Background

I implemented idempotency for my Go Banking API to ensure that duplicate client requests do not execute the same operation multiple times.

The expected behavior was:

* The first request executes the transfer.
* Subsequent requests with the same `Idempotency-Key` return the previously stored response.
* The transfer should execute only once.

This worked correctly for sequential requests.

---

# Initial Flow

The implementation followed this logic:

```text
Receive Request
      │
      ▼
Lookup Idempotency Key
      │
      ├── Exists
      │      ▼
      │  Return Stored Response
      │
      └── Not Found
             ▼
      Execute Transfer
             ▼
      Insert Idempotency Record
             ▼
        Return Response
```

At first glance this looked correct.

---

# How I Found the Bug

I suspected there might be problems under concurrent requests.

To verify this, I wrote a small Go program that sent multiple HTTP requests simultaneously using the same `Idempotency-Key`.

Example:

```go
var wg sync.WaitGroup

for i := 0; i < 20; i++ {
    wg.Add(1)

    go func() {
        defer wg.Done()

        // POST /transfer
        // Same Idempotency-Key
    }()
}

wg.Wait()
```

All requests used

```
Idempotency-Key: abc123
```

---

# Expected Result

20 requests

↓

Only one transfer should execute.

```
Account A: 1000 -> 900
Account B: 1000 -> 1100
```

Every request should receive the same response.

---

# Actual Result

Some requests returned

```
HTTP 200
```

Others returned

```
duplicate key value violates unique constraint
```

More importantly,

before test

```
Account A = 1000
Account B = 1000
```

after test

```
Account A = 400
Account B = 1600
```

This proved that six transfers executed instead of one.

The idempotency implementation failed under concurrency.

---

# Root Cause

The implementation had a classic **check-then-act race condition**.

Multiple goroutines executed:

```
Check if key exists
```

at almost the same time.

Every request observed

```
Key not found
```

before any request inserted the idempotency record.

The timeline looked like this:

```
Request A
Request B
Request C

↓

Lookup key

↓

Not Found
Not Found
Not Found

↓

Execute Transfer
Execute Transfer
Execute Transfer

↓

Insert Key
```

Only one insert succeeded.

The others failed because of the unique constraint.

However, the money had already been transferred multiple times.

---
# Code
```
     idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	existing, err := s.idempotencyStore.GetIdempotency(ctx, userID, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var response dto.TransferResponse
		if err := json.Unmarshal(existing.Response, &response); err != nil {
			return nil, err
		}
		return &response, nil
	}

	sender, err := s.accStore.GetAccountByID(ctx, req.FromID)
	if err != nil {
		return nil, fmt.Errorf("sender: %w", err)
	}

	if sender.UserID != userID {
		return nil, errors.New("forbidden")
	}

	if _, err := s.accStore.GetAccountByID(ctx, req.ToID); err != nil {
		return nil, fmt.Errorf("receiver: %w", err)
	}

	if sender.Balance < req.Amount {
		return nil, fmt.Errorf("insufficient balance: have %d need %d", sender.Balance, req.Amount)
	}

	if err := s.accStore.TransferTx(ctx, req.FromID, req.ToID, req.Amount); err != nil {
		return nil, err
	}
```
---
# Why the Unique Constraint Wasn't Enough

The database correctly prevented duplicate rows inside the `idempotency_keys` table.

However,

the business operation had already completed before the insert happened.

Therefore,

the unique constraint protected only the idempotency table,

not the money transfer itself.

---

# How I Verified the Bug

I checked the balances after the concurrent test.

Instead of

```
900
1100
```

I observed

```
400
1600
```

which proved multiple transfers were committed.

---

# Lessons Learned

Sequential testing is not enough.

A feature that appears correct with

```
Request 1
Request 2
Request 3
```

may completely fail under concurrent execution.

Concurrency testing exposed a production-level bug.

---

# Next Improvement

The idempotency key must be claimed **before** executing the transfer.

The entire workflow should be protected so that only one request can perform the business operation.

High-level idea:

```
Receive Request
      │
      ▼
Atomically claim idempotency key
      │
      ├── Already claimed
      │      ▼
      │ Return stored response
      │
      └── Successfully claimed
             ▼
        Execute Transfer
             ▼
        Store Response
             ▼
          Commit
```

This guarantees that duplicate concurrent requests never execute the transfer multiple times.

---

# Additional Improvements

Future improvements include:

* Scope idempotency keys per operation (e.g. transfer, deposit, withdraw).
* Store a hash of the original request and reject reuse of the same key with a different payload.
* Add integration tests for concurrent requests.
* Add load tests to validate idempotency under heavy traffic.
* Return the cached response instead of exposing database duplicate-key errors to clients.

---

# Key Takeaway

Idempotency is not only about preventing duplicate requests.

It must also remain correct under concurrent execution.

A concurrency test uncovered a race condition that normal sequential testing could not detect, highlighting the importance of testing backend systems under realistic production scenarios.
