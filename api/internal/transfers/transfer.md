# Transfer API
The API exposes two endpoints:

## Get Transfers
GET **/transfer**

Require authentication in the form of a Authorization header with a valid token.

Return a list of all transaction involving the logged in user of the form :

```
[
  {
    "source_id": 505,
    "source": "Alice",
    "destination_id": 506,
    "destination": "Bob",
    "amount": 10,
    "execution_date": "2026-10-26T13:20:01.778864Z"
  },
  {
    "source_id": 506,
    "source": "Bob",
    "destination_id": 507,
    "destination": "Charlie",
    "amount": 5,
    "execution_date": "2026-10-26T13:24:41.485474Z"
  },
  {
    "source_id": 506,
    "source": "Bob",
    "destination_id": 507,
    "destination": "Charlie",
    "amount": 10,
    "execution_date": "2026-10-26T13:24:49.29277Z"
  },
  {
    "source_id": 506,
    "source": "Bob",
    "destination_id": 505,
    "destination": "Alice",
    "amount": 3,
    "execution_date": "2026-10-26T13:37:20.800943Z"
  },
  {
    "source_id": 506,
    "source": "Bob",
    "destination_id": 505,
    "destination": "Alice",
    "amount": 4,
    "execution_date": "2026-10-26T13:37:52.86031Z"
  }
]
```
source_id is the internal ID of the source of the transfer
source is the username of the source of the transfer

destination_id is the internal ID of the destination of the transfer
destination is the username of the destination of the transfer

amount is the quantity of currency transfered

execution_date is the server time when the transfer was done

## Create Transfer
POST **/transfer**

Require authentication in the form of a Authorization header with a valid token.

Require a payload of the form :

```
{
  "destination_id":506,
  "amount":4
}
```

destination_id is the internal ID of the destination of the transfer. It must be the ID of an existing user.

amount is the amount of currency to transfer. Amount must be strictly positive, and less than the current balance of the logged in user.

Return the new transaction:

```
{
  "source_id": 505,
  "source": "Alice",
  "destination_id": 506,
  "destination": "Bob",
  "amount": 4,
  "execution_date": "2026-10-26T13:20:01.778864Z"
}
```
Same keys as above
