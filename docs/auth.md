# Auth Service Execution Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         AUTH SERVICE                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐               │
│  │   Client     │─────▶│ gRPC Server  │────▶ │   Service    │               │
│  │  (Gateway)   │      │ (Controller) │      │  (Logic)     │               │
│  └──────────────┘      └──────────────┘      └──────────────┘               │
│                                                      │                      │
│                      ┌──────────────┐                │                      │
│                      │   JWT Mgr    │◀───────────────┤                      │
│                      │ (Token Gen)  │                │                      │
│                      └──────────────┘                │                      │
│                                                      │                      │
│                      ┌──────────────┐      ┌──────--─┴───────┐              │
│                      │   Redis      │◀─────│   PostgreSQL    │              │
│                      │ (Sessions)   │      │   (Users DB)    │              │
│                      └──────────────┘      └───────────────-─┘              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────--------┘

┌─────────────────────────────────────────────────────────────────────┐
│                        ENDPOINTS                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. REGISTER ──┐                                                    │
│     │          │  Input: email, password, fullName                  │
│     ▼          │  1. Validate email format + password strength      │
│  Check:        │  2. Check if email exists                          │
│  existing      │  3. Hash password (bcrypt cost=12)                 │
│  user?         │  4. Create user in PostgreSQL                      │
│     │          │  Return: userID, email                             │
│     No ──▶ Create                                                   │
│                                                                     │
│  2. LOGIN ──────┐                                                   │
│     │           │  Input: email, password                           │
│     ▼           │  1. Find user by email                            │
│  Find user      │  2. Verify password with bcrypt                   │
│     │           │  3. Generate access token (15min TTL)             │
│     ▼           │  4. Generate refresh token (7 days)               │
│  Verify pass    │  5. Store refresh token in Redis                  │
│     │           │  Return: accessToken, refreshToken                │
│     Valid ──▶ Issue                                                 │
│                                                                     │
│  3. VALIDATE_TOKEN ──▶                                              │
│     Input: access_token                                             │
│     1. Parse JWT (HS256)                                            │
│     2. Validate signature + expiration + issuer                     │
│     Return: userID, email, valid=true/false                         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Core Components

| Component | File | Role |
|-----------|------|------|
| main.go | cmd/authsvc/ | Entry point, server setup, graceful shutdown |
| gRPC_server.go | internal/auth | Controller - validates input, maps request to service |
| service.go | internal/auth | Core logic - business rules |
| jwt.go | internal/auth | Token generation & validation |
| password.go | internal/auth | Password hashing (bcrypt) |

## Flow Summary

```
Register: Client → Validate inputs → Check DB → Hash pass → Save DB → Return user
Login:    Client → Find user → Verify pass → Gen JWTs → Store Redis → Return tokens  
Validate: Client → Parse JWT → Verify signature + exp → Return claims
```