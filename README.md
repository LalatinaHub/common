# LalatinaHub/common

[![Go Reference](https://pkg.go.dev/badge/github.com/LalatinaHub/common.svg)](https://pkg.go.dev/github.com/LalatinaHub/common)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`common`** is the foundation (Leaf Module) for the LalatinaHub ecosystem. It provides core domain models, Turso LibSQL database connection pooling, repository interfaces and SQL implementations, bi-directional proxy URL parsers/formatters, pure-Go multi-format client subconverters (Clash Meta YAML, sing-box v1.14+ JSON, SFA/BFR, Base64), IATA airport geolocation lookup, network probe runners, and high-performance UDP packet relaying.

Designed strictly following the **Leaf Module Pattern**:
- 🟢 **Zero heavy runtime dependencies**: Only relies on the Go standard library, `libsql-client-go`, and `yaml.v3`.
- 🟢 **No Framework Lock-in**: Zero imports of `sing-box`, `caddy`, `gin`, or heavy proxy engines.
- 🟢 **Cross-Project Reusability**: Safe for import across `LatinaServer`, `LatinaApi` (subscription gateway), `LatinaBot` (Telegram Bot), CLI utilities, and Web Dashboards without dependency version conflicts.

---

## 📦 Packages Overview

```
github.com/LalatinaHub/common
├── database/          # Turso LibSQL connection pool & index management
├── model/             # Core domain models (User, Server, KeyValue, ProxyNode)
├── proxy/             # Bi-directional proxy URL parser & formatter (SS, VMess, VLESS, Trojan)
├── repository/        # SQL repositories and data access interfaces
│   └── mocks/         # Testify mocks for unit tests
├── subconverter/      # Pure-Go client profile generator (Clash Meta YAML, sing-box v1.14+ JSON, Base64, SFA/BFR)
├── region/            # IATA 3-letter airport code geolocation lookup table (9,200+ entries, O(1))
├── probe/             # Concurrent network diagnostic probes (YouTube CDN, Netflix unlock)
└── udprelay/          # High-performance UDP packet relay with sync.Pool & timeout safety
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
Bi-directional, pure-Go parser and formatter for popular proxy URL formats:
- **`Parser`**: Parses `ss://`, `vmess://`, `vless://`, and `trojan://` URLs into `*model.ProxyNode`.
- **`Formatter`**: Serializes `*model.ProxyNode` into standard proxy URLs (`Format` and `FormatString`).

### 5. `subconverter`
Universal client configuration converter without importing heavy proxy core runtimes:
- **Clash Meta (Mihomo)**: Generates complete YAML with proxy groups (`PROXIES`, `AUTO-FALLBACK`, `LOAD-BALANCE`) and rules.
- **sing-box v1.14+**: Generates modern JSON configurations with DNS, Inbounds, Outbounds, and Rules.
- **SFA / BFR**: Compatibility profiles for mobile sing-box clients.
- **Raw & Base64**: Plain-text lines and standard Base64-encoded strings for v2rayNG, Shadowrocket, and NekoBox.
- **Template Post-processing**: Supports `"cf"` Cloudflare Trojan UDP detour injection.

### 6. `region`
High-speed geolocation lookup from 3-letter IATA airport codes:
- Pre-compiled static mapping table with over 9,200 international airport codes.
- `Lookup(code string) (string, bool)` with case-insensitive, zero-allocation lookup.

### 7. `probe`
Network diagnostic probe engine:
- Concurrent worker execution with panic recovery and context deadline enforcement.
- **`YouTubeCDN`**: Extracts IATA airport code and city mapping via Google Video redirector.
- **`Netflix`**: Detects streaming license unlocking and country code.

### 8. `udprelay`
Zero-dependency UDP packet relay:
- Thread-safe buffer pooling with `sync.Pool` (2048-byte buffers).
- Direct raw UDP relay (`Relay`) and Base64-encoded HTTP-friendly relay (`RelayBase64`).
- Strict context timeout and deadline management.

---

## 🚀 Installation

```bash
go get github.com/LalatinaHub/common@latest
```

---

## 💡 Usage Examples

### Parsing and Formatting Proxy URLs

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

	// Re-serialize back to URL
	urlStr, err := proxy.FormatString(node)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Formatted URL: %s\n", urlStr)
}
```

### Converting Proxy Nodes to Clash and sing-box

```go
package main

import (
	"fmt"
	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/subconverter"
)

func main() {
	nodes := []model.ProxyNode{
		{
			VPN:        "trojan",
			Server:     "tr.example.com",
			ServerPort: 443,
			Password:   "secret",
			TLS:        true,
			Remark:     "Trojan-SG",
		},
	}

	conv := subconverter.New(nodes)

	// Generate Clash Meta YAML
	clashYAML, _ := conv.ToClash("cf")
	fmt.Println(clashYAML)

	// Generate sing-box v1.14+ JSON
	singboxJSON, _ := conv.ToSingbox("standard", "cf")
	fmt.Println(singboxJSON)

	// Generate Base64 subscription
	b64 := conv.ToBase64()
	fmt.Println(b64)
}
```

### Checking Edge Node Geolocation with IATA

```go
package main

import (
	"fmt"
	"github.com/LalatinaHub/common/region"
)

func main() {
	if city, ok := region.Lookup("CGK"); ok {
		fmt.Printf("Airport CGK is in %s\n", city) // JAKARTA-CENGKARENG
	}
}
```

---

## 🧪 Testing

Run test suite:

```bash
go test -v ./...
```

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
