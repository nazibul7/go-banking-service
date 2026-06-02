# Partial Signup Causing Inconsistency Data

While testing my Go + PostgreSQL banking API, I hit a real production-style issue during user creation.

## Where it happened

In the `/signup` endpoint.

## How I reproduced it

* Register a new user
* Temporarily inject an error in generateRefreshToken function.
* Got error in refresh token creation.
* Inspect the DB, found user data is available
---

# Root Cause

User creation & refresh token persistence were executed as seperate database operations without a transaction boundary.

---

# The fix

Wrap user creation & refresh token persistenc inside a single db transaction.

Transaction flow:
* BEGIN
* INSERT user
* INSERT refresh_token
* COMMIT

If any step fails:
* ROLLBACK

---

# Biggest learning

When multiple db operations represent a single business operation, they should succeed or fail together. Transactions preserve consistency & prevent partial writres.