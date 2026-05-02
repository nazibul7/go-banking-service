# PostgreSQL & Go — Things I Learned in Banking App Project

## Transactions in Go

* Transactions are useful when multiple database changes depend on each other and must either:

  * all succeed together, or
  * all fail together.

* Example:

  * A money transfer requires:

    1. deducting balance from sender
    2. adding balance to receiver
    3. recording transaction history

  If any step fails, the entire operation should rollback.

* Simple single-row operations like a basic deposit or withdraw may not always require an explicit transaction if a single SQL statement can guarantee correctness atomically.

* In Go, transactions usually follow this pattern:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}

defer tx.Rollback()

// queries using tx.Exec / tx.Query

if err := tx.Commit(); err != nil {
    return err
}
```

### Important Transaction Behavior

* `defer tx.Rollback()` always runs when the function returns.

* If `tx.Commit()` succeeds:

  * PostgreSQL marks the transaction as completed.
  * The deferred rollback becomes a no-op (it does nothing).

* Using deferred rollback is a common and safe production pattern because it guarantees cleanup if:

  * an error occurs
  * the function returns early
  * a panic happens before commit
