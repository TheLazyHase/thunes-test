# User API
The API exposes two endpoints:

## Login
POST **/login**, which is expecting a POST request with a JSON payload of the form:

Require a payload of the form :
```
{"username": "Alice", "password": "password123"}
```

If the credentials are valid, it responds with 200 OK, returning a JWT token in the response body which is valid for 60 minutes.

## My informations
GET **/me**

Require authentication in the form of a Authorization header with a valid token.

Return relevant informations about the currently logged in user.

```
{
  "username":"Bob",
  "balance":87,
  "id":506
}
```
username is the name of the logged in user
id is the internal ID of the logged in user

balance is the amount of currency the user have

## Get users
GET **/users**

Require authentication in the form of a Authorization header with a valid token.

Return the list of all users:

```
[
  {
    "username": "Charlie",
    "id": 507
  },
  {
    "username": "Alice",
    "id": 505
  },
  {
    "username": "Bob",
    "id": 506
  },
  {
    "username": "David",
    "id": 508
  }
]
```

id is the internal ID of the user.

username is the name of the user
