# Negative Balance

## What I had
```go
query := `UPDATE accounts SET balance = balance + $1 WHERE id=$2`
```
No rule — just math. Balance could go to -∞ and the db would accept it happily.

## What I thought the problem was
10 concurrent withdrawals 
```
 seq 1 10 | xargs -P10 -I{} curl -X PATCH http://localhost:8080/account/1/withdraw -H "Content-Type: application/json" -d '{"amount":200}'
```
→ balance hit -1200 → assumed race condition( though about adding transaction but then realize it was a atomic query, it has to be business logic not race or concurrent issue)

## What the actual problem was
Not a concurrency issue. Not a race condition.
Pure missing business logic — never told the system balance cannot go negative.

Even a single request would do it:
```bash
curl -X PATCH http://localhost:8080/account/1/withdraw -d '{"amount": 99999}'
```
No concurrency needed. Balance goes negative instantly.

## The fix
```go
// withdraw — guard lives in SQL, not in Go code
query = `UPDATE accounts SET balance = balance + $1 
         WHERE id = $2 
         AND balance + $1 >= 0`

// deposit — no guard needed, adding money always valid
query = `UPDATE accounts SET balance = balance + $1 
         WHERE id = $2`
```

## Why SQL and not Go
Checking balance in Go = read then write = two steps = needs transaction
Checking balance in SQL = one atomic statement = no transaction needed

## How to tell the difference next time
| Symptom | Problem |
|---|---|
| Happens with single request too | Business logic issue |
| Only happens with concurrent requests | Race condition |

## What rows=0 means now
Either account id does not exist OR insufficient funds.
Return different errors for each if you need to tell them apart.