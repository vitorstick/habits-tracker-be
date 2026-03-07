# Habit Tracker API Usage Guide

This guide explains how to use the Habit Tracker backend as a standalone API, suitable for scripts, CLI tools, or other applications.

## 🚀 Quick Start

1. **Start the Backend**: Ensure your Go server is running (usually on `http://localhost:8080`).
2. **Production URL**: The API is also available at `https://habits-tracker-be.onrender.com/api/`.
3. **Authentication**: All protected routes require a JWT (JSON Web Token) in the `Authorization` header.


---

## 🔐 Authentication Flow

1. **Login or Register**: Send a POST request to `/api/auth/login` or `/api/auth/register`.
2. **Retrieve Token**: The response will contain a `token` field.
3. **Use Token**: Include this token in subsequent requests:
   ```http
   Authorization: Bearer <your_token_here>
   ```

---

## 🛠️ Usage Examples

### Using `curl`

**Login:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com", "password":"password123"}'
```

**Fetch Habits:**
```bash
curl -X GET http://localhost:8080/api/habits \
     -H "Authorization: Bearer <TOKEN>"
```

---

## 🐍 Python Scripting

We've provided a sample client in `scripts/api_client.py`. You can use it as a base for your own scripts.

```python
import requests

BASE_URL = "http://localhost:8080/api"
header = {"Authorization": "Bearer <TOKEN>"}

# example: fetch habits
response = requests.get(f"{BASE_URL}/habits", headers=header)
print(response.json())
```

---

## 🧪 Testing with REST Client
If you use VS Code, you can use the `api_tests.http` file with the **REST Client** extension to quickly test and inspect all endpoints.
