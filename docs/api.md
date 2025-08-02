# API Documentation

## POST /check-access

**Request:**

```json

{
  "subject": "user1",
  "resource": "file1",
  "action": "read",
  "conditions": {}
}
```

**Response:**

```json
{
  "allow": true,
  "policy_id": "policy1",
  "reason": "allowed by policy",
  "context": {
    "subject": "user1",
    "resource": "file1",
    "action": "read"
  }
}
```
