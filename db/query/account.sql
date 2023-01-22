-- name: CreateAccount :one
INSERT INTO accounts (
                      owner,
                      balance,
                      currency
) VALUES (
          $1, $2, $3
         ) RETURNING *;
/*
 This GetAccount query is a simple
 select. So it will not stop another query to access the same
 account while the first one is still running.
 This not a desirable behaviour. Thus we will
 implement a way to have a lock until the first query is over
 before other query can access that account

 */

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1
LIMIT 1;

/*
 So this GetAccountForUpdate will solve the problem encountered with
 GetAccount, which is to lock create a lock until the transaction is committed

 the transaction lock is only required because Postgres worries that transaction
 will update the account ID which will affect the foreign key constraints of transfers table.

 Since updateAccount query only change the account balance and also since the account ID
 is the primary key, it will never change

 Thus we use 'FOR NO KEY UPDATE' to tell Postgres that we are selecting this account for update
 but the primary key won't be touched. This way, Postgres will not need to acquire the transaction
 lock. Thus we no longer have the deadlock issue
 */

-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListAccounts :many
SELECT * FROM accounts
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;


/*
 Currently we have to run 2 queries to get the account and update its balance
 we can improve that by simply running one query

 this query set directly the value of the account to the new value by directly
 adding or removing the amount being transferred

 we use sqlc.arg(value) to specify the argument name to value

 */

-- name: AddAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;