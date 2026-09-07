# knock (Go Refactored Client)

*Read this in other languages:* **English** | [Português (Brasil)](README-PTBR.md)

---

A modern, cross-platform, highly testable **port-knocking** client, refactored in **Go** from Judd Vinet's original C implementation.

The original `knock` project was built for Unix/Linux systems with C library dependencies. This fork refactors the **client application** (`knock`) into pure Go (zero CGo), making it natively buildable and runnable on **Windows**, **Linux**, and **macOS** (Intel and Apple Silicon), while strictly preserving compatibility with the original client's flags and behavior.

---

## 🚀 Features of the Go Client

- **Native Cross-Platform**: Effortless compilation for Windows (`.exe`), Linux, and macOS (`amd64` and `arm64`) with no external DLL dependencies or CGo.
- **TCP and UDP Protocols**:
  - **TCP**: Non-blocking TCP SYN packet dispatch (short 200ms timeout).
  - **UDP**: 1-byte datagram payload matching the original specification.
  - Global flag support (`-u` to default all hits to UDP) or per-port protocol overrides (`port:proto`, e.g., `7000:tcp`, `8000:udp`).
- **IPv4 / IPv6 Resolution**:
  - `-4/--ipv4` flag to force IPv4 resolution.
  - `-6/--ipv6` flag to force IPv6 resolution.
  - Default mode supporting automatic DNS resolution and direct IP addresses.
- **Interval Control (Delay)**:
  - `-d/--delay <t>` flag to wait `<t>` milliseconds between consecutive port hits.
- **Verbose Output**:
  - `-v/--verbose` flag formatted identically to `knock.c`:
    ```text
    hitting tcp 192.168.1.1:7000
    hitting udp 192.168.1.1:8000
    ```
- **Clean Architecture and Design Patterns**:
  - **Domain-Driven Design (DDD)**: Value Objects (`KnockTarget`, `Port`, `Protocol`, `IPVersion`) and Aggregate Root (`KnockExecution`).
  - **Ports and Adapters (Hexagonal Architecture)**: Decoupled interfaces (`ports.Knocker`, `ports.Sleeper`, `ports.Printer`) isolating network and console I/O.
  - **Builder Pattern**: `KnockSequenceBuilder` for robust sequence construction and validation.
  - **Strategy Pattern**: `TCPHitStrategy` and `UDPHitStrategy` encapsulating protocol-specific network logic.
  - **Decorator Pattern**: `LoggingKnockerDecorator` for latency measurement and telemetry.
  - **Observer & Message Bus**: Decoupled message bus for command dispatching and event notification.
  - **Test Doubles (Fakes)**: `FakeKnocker` and `FakeSleeper` enabling instant unit and service test execution in milliseconds.

---

## 🛠️ How to Build

### Prerequisites
- **Go 1.22** or higher installed on your system.

### Get the Project
```bash
git clone https://github.com/ph-py/knock-go.git
cd knock-go
```

Or install the binary directly using:
```bash
go install github.com/ph-py/knock-go/cmd/knock@latest
```

### Native Compilation

#### On Windows (PowerShell / CMD)
```powershell
go build -o bin/knock.exe ./cmd/knock
```

#### On Linux / macOS
```bash
go build -o bin/knock ./cmd/knock
```

---

### Cross-Compilation

Since the Go client does not use CGo (`CGO_ENABLED=0` by default), you can generate binaries for any target OS from any platform:

#### Build for Windows (64-bit):
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/knock.exe ./cmd/knock
```

#### Build for Linux (64-bit):
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/knock-linux-amd64 ./cmd/knock
```

#### Build for macOS Intel:
```powershell
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o bin/knock-darwin-amd64 ./cmd/knock
```

#### Build for macOS Apple Silicon (M1/M2/M3/M4):
```powershell
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/knock-darwin-arm64 ./cmd/knock
```

---

## 🧪 Running Tests

Run the full test suite covering domain tests (Low Gear) and service orchestration (High Gear):

```powershell
go test -v ./...
```

---

## 📖 Usage

### Syntax
```bash
knock [options] <host> <port[:proto]> [port[:proto]] ...
```

### Options
| Option | Description |
| :--- | :--- |
| `-u, --udp` | Use UDP as default protocol for all ports (default is TCP) |
| `-d, --delay <ms>` | Wait `<ms>` milliseconds between each port hit |
| `-4, --ipv4` | Force IPv4 resolution and connection |
| `-6, --ipv6` | Force IPv6 resolution and connection |
| `-v, --verbose` | Show verbose status messages for each hit |
| `-V, --version` | Display program version |
| `-h, --help` | Display help screen |

### Examples

1. **Mixed TCP and UDP sequence with verbose output:**
   ```bash
   knock myserver.example.com 7000:tcp 8000:udp 9000:tcp -v
   ```

2. **Set a 100ms delay between port hits:**
   ```bash
   knock myserver.example.com 1111 2222 3333 -d 100 -v
   ```

3. **Force all ports to use UDP:**
   ```bash
   knock myserver.example.com 1234 5678 -u -v
   ```

4. **Force IPv4 resolution:**
   ```bash
   knock -4 myserver.example.com 7000 8000 9000
   ```

---

## ⚖️ License and Credits

### Original Project
This project is based on the **knock/knockd** port-knocking utility originally created by:
- **Judd Vinet** (`jvinet@zeroflux.org`) — Original author and maintainer.

With contributions over the years by (as listed in `CONTRIBUTERS`):
- airwoflgh <paul.rogers@flumps.org>
- catbref <misc-github@talk2dom.com>
- Diego Elio Pettenò <flameeyes@flameeyes.eu>
- Dima Krasner <dima@dimakrasner.com>
- Jonathon Reinhart <jonathon.reinhart@gmail.com>
- Marius Hoch <hoo@online.de>
- Michael Weiss <dev.primeos@gmail.com>
- Oswald Buddenhagen <ossi@kde.org>
- Sébastien Valat <sebastien.valat@gmail.com>
- TDFKAOlli <TDFKAOlli@ish.de>
- Ximin Luo <infinity0@pwned.gg>
- vriera <Vincent.Riera@imgtec.com>
- And open source community contributors.

### Fork Licensing
The original code is licensed under the **GNU General Public License v2.0 or later** ([COPYING](COPYING)).

As a derivative work and refactoring of original GPLv2+ code, this fork is distributed under the **GNU General Public License v2.0** (or **GPLv3**, as permitted by the *"either version 2 of the License, or (at your option) any later version"* clause).
