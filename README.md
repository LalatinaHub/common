# LalatinaHub/common

[![Go Reference](https://pkg.go.dev/badge/github.com/LalatinaHub/common.svg)](https://pkg.go.dev/github.com/LalatinaHub/common)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`common`** is the foundation (Leaf Module) for the LalatinaHub ecosystem. It provides core domain models, Turso LibSQL database connection pooling, repository interfaces and SQL implementations, and generic proxy URL parsers.

Designed strictly following the **Leaf Module Pattern**:
- 🟢 **Zero heavy dependencies**: Only relies on the Go standard library and the official `libsql-client-go` driver.
- 🟢 **No Framework Lock-in**: Zero imports of `sing-box`, `caddy`, `gin`, or other runtime engines.
- 🟢 **Cross-Project Reusability**: Safe for import across `LatinaServer`, `LatinaBot` (Telegram Bot), CLI utilities, and Web Dashboards without dependency version conflicts.

---

## 📦 Packages Overview

```
github.com/LalatinaHub/common
├── database/          # Turso LibSQL connection pool & index management
├── model/             # Core domain models (User, Server, KeyValue, ProxyNode)
├── proxy/             # Generic proxy URL parser (SS, VMess, VLESS, Trojan)
└── repository/        # SQL repositories and data access interfaces
    └── mocks/         # Testify mocks for unit tests
```

### 1. `model`
Domain entities representing core business models:
- **`User`**: VPN user with expiration, server code, and byte quota. Contains `IsActive(now)` method.
- **`Server`**: Cluster edge node server with capacity checking (`IsFull()`).
- **`KeyValue`**: Generic configuration key-value pair.
- **`ProxyNode`**: Universal proxy definition supporting Shadowsocks, VMess, VLESS, and Trojan.

### 2. `database`
Connection management and pooling for Turso LibSQL:
- Singleton connection pool configured for resilience and concurrency (`MaxOpenConns: 25`, `MaxIdleConns: 5`).
- Environment variable configuration (`TURSO_DATABASE_URL`, `TURSO_AUTH_TOKEN`).
- `InitIndexes(ctx)` to ensure essential performance indexes exist.
- Helper methods: `Ping`, `Close`, `SetDB`, `ResetDBInstance`.

### 3. `repository`
Clean architecture repository implementations:
- **`UserRepository`**: `GetActiveUsersGroupedByVPN`, `DeductQuota`, `DeductQuotaBatch` (transactional), `CreateUser`.
- **`ServerRepository`**: `GetAll`.
- **`KVRepository`**: `GetAll`.
- **`ProxyRepository`**: `GetRelays`.
- **`mocks`**: Pre-generated `testify/mock` structs for unit testing consumer services.

### 4. `proxy`
Pure Go parser for popular proxy URL formats:
- Supports `ss://` (Shadowsocks with base64 and standard URI formats).
- Supports `vmess://` (Base64 JSON format).
- Supports `vless://` (VLESS with query parameters: ws, grpc, tls).
- Supports `trojan://` (Trojan with query parameters: ws, tls).

---

## 🚀 Installation

```bash
go get github.com/LalatinaHub/common@v0.1.0
```

---

## 💡 Usage Examples

### Parsing Proxy URLs

```go
package main

import (
	"fmt"
	"github.com/LalatinaHub/common/proxy"
)

func main() {
	p := proxy.NewParser()
	node, err := p.Parse("ss://YWVzLTEyOC1nY206cGFzc3dvcmRAZXhhbXBsZS5jb206ODM4OA==#Example")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Server: %s:%d, Method: %s\n", node.Server, node.ServerPort, node.Method)
}
```

### Initializing Database and Repositories

```go
package main

import (
	"context"
	"fmt"
	"github.com/LalatinaHub/common/database"
	"github.com/LalatinaHub/common/repository"
)

func main() {
	db, err := database.GetDB()
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUserRepository(db)
	users, err := userRepo.GetActiveUsersGroupedByVPN(context.Background())
	if err != nil {
		panic(err)
	}

	for vpn, list := range users {
		fmt.Printf("[%s] Active users: %d\n", vpn, len(list))
	}
}
```

### Using Repository Mocks in Tests

```go
package mytest

import (
	"context"
	"testing"
	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMyService(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	ctx := context.Background()

	mockUserRepo.On("GetActiveUsersGroupedByVPN", ctx).Return(map[string][]model.User{
		"trojan": {{ID: 1, Token: "abc"}},
	}, nil)

	users, err := mockUserRepo.GetActiveUsersGroupedByVPN(ctx)
	assert.NoError(t, err)
	assert.Len(t, users["trojan"], 1)
}
```

---

## 🧪 Testing

Run test suite with coverage:

```bash
go test -v -race -cover ./...
```

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
