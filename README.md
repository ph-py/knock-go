# knock (Go Refactored Client)

Um cliente moderno, multiplataforma e de alta testabilidade para **port-knocking**, refatorado em **Go** a partir da implementação original em C de Judd Vinet.

O projeto original `knock` foi desenvolvido para sistemas Unix/Linux com dependência de bibliotecas C. Este fork refatora a **aplicação cliente** (`knock`) para Go puro (sem CGo), tornando-a nativamente compilável e executável em **Windows**, **Linux** e **macOS** (Intel e Apple Silicon), mantendo estrita compatibilidade com os parâmetros e comportamento do cliente original.

---

## 🚀 Funcionalidades do Cliente Go

- **Multiplataforma Nativo**: Compilação sem esforço para Windows (`.exe`), Linux e macOS (`amd64` e `arm64`) sem dependência de DLLs externas ou CGo.
- **Protocolos TCP e UDP**:
  - **TCP**: Disparo de pacote TCP SYN não-bloqueante (timeout curto de 200ms).
  - **UDP**: Envio de datagrama de 1 byte compatível com a especificação original.
  - Suporte a flag global `-u` (UDP como padrão) ou definição individual por porta (`porta:proto`, ex: `7000:tcp`, `8000:udp`).
- **Resolução IPv4 / IPv6**:
  - Suporte a `-4/--ipv4` para forçar IPv4.
  - Suporte a `-6/--ipv6` para forçar IPv6.
  - Modo padrão com resolução automática de DNS ou uso de IPs diretos.
- **Controle de Intervalo (Delay)**:
  - Flag `-d/--delay <t>` para aguardar `<t>` milissegundos entre batidas consecutivas.
- **Saída Detalhada (Verbose)**:
  - Flag `-v/--verbose` com formatação idêntica à do `knock.c`:
    ```text
    hitting tcp 192.168.1.1:7000
    hitting udp 192.168.1.1:8000
    ```
- **Arquitetura Limpa e Padrões de Projeto**:
  - **Domain-Driven Design (DDD)**: Value Objects (`KnockTarget`, `Port`, `Protocol`, `IPVersion`) e Aggregate Root (`KnockExecution`).
  - **Ports and Adapters (Hexagonal Architecture)**: Interfaces desacopladas (`ports.Knocker`, `ports.Sleeper`, `ports.Printer`) isolando I/O de rede e console.
  - **Builder Pattern**: `KnockSequenceBuilder` para construção e validação robusta de sequências.
  - **Strategy Pattern**: `TCPHitStrategy` e `UDPHitStrategy` encapsulando a lógica específica de cada protocolo.
  - **Decorator Pattern**: `LoggingKnockerDecorator` para medição de latência e telemetria.
  - **Observer & Message Bus**: Barramento desacoplado para despacho de comandos e notificação de eventos.
  - **Test Doubles (Fakes)**: `FakeKnocker` e `FakeSleeper` permitindo testes unitários e de serviço instantâneos em milissegundos.

---

## 🛠️ Como Construir (Build)

### Pré-requisitos
- **Go 1.22** ou superior instalado no sistema.

### Compilação Nativa

#### No Windows (PowerShell / CMD)
```powershell
go build -o bin/knock.exe ./cmd/knock
```

#### No Linux / macOS
```bash
go build -o bin/knock ./cmd/knock
```

---

### Compilação Cruzada (Cross-Compilation)

Como o cliente Go não utiliza CGo (`CGO_ENABLED=0` por padrão), você pode gerar binários para qualquer sistema operacional a partir de qualquer plataforma:

#### Compilar para Windows (64-bit):
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/knock.exe ./cmd/knock
```

#### Compilar para Linux (64-bit):
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/knock-linux-amd64 ./cmd/knock
```

#### Compilar para macOS Intel:
```powershell
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o bin/knock-darwin-amd64 ./cmd/knock
```

#### Compilar para macOS Apple Silicon (M1/M2/M3/M4):
```powershell
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/knock-darwin-arm64 ./cmd/knock
```

---

## 🧪 Executando os Testes

Execute a suíte de testes com testes de domínio (Low Gear) e orquestração de serviço (High Gear):

```powershell
go test -v ./...
```

---

## 📖 Como Usar

### Sintaxe
```bash
knock [options] <host> <port[:proto]> [port[:proto]] ...
```

### Opções
| Opção | Descrição |
| :--- | :--- |
| `-u, --udp` | Usa UDP como protocolo padrão para todas as portas (o padrão é TCP) |
| `-d, --delay <ms>` | Tempo de espera em milissegundos entre cada batida de porta |
| `-4, --ipv4` | Força a resolução e conexão usando IPv4 |
| `-6, --ipv6` | Força a resolução e conexão usando IPv6 |
| `-v, --verbose` | Exibe mensagens detalhadas de cada porta batida |
| `-V, --version` | Exibe a versão do programa |
| `-h, --help` | Exibe a tela de ajuda |

### Exemplos

1. **Sequência mista de TCP e UDP com verbose:**
   ```bash
   knock myserver.example.com 7000:tcp 8000:udp 9000:tcp -v
   ```

2. **Definir delay de 100ms entre as batidas:**
   ```bash
   knock myserver.example.com 1111 2222 3333 -d 100 -v
   ```

3. **Forçar todas as portas para UDP:**
   ```bash
   knock myserver.example.com 1234 5678 -u -v
   ```

4. **Forçar uso de IPv4:**
   ```bash
   knock -4 myserver.example.com 7000 8000 9000
   ```

---

## ⚖️ Licença e Créditos

### Projeto Original
Este projeto é baseado no utilitário de port-knocking **knock/knockd** criado por:
- **Judd Vinet** (`jvinet@zeroflux.org`) — Autor e mantenedor original.

Com contribuições ao longo dos anos por (conforme arquivo `CONTRIBUTERS`):
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
- E os colaboradores da comunidade open source.

### Licenciamento do Fork
O código original está sob a licença **GNU General Public License v2.0 or later** ([COPYING](file:///c:/Users/ph_ol/source/repos/knock-win/knock-win/COPYING)).

Como este projeto é um trabalho derivado / refatoração de código original sob GPLv2+, o seu fork deve ser distribuído sob a **GNU General Public License v2.0** (ou **GPLv3**, conforme permitido pela cláusula *"either version 2 of the License, or (at your option) any later version"*).

Ao criar o seu fork:
1. Mantenha o arquivo `COPYING` existente.
2. Adicione sua identificação de copyright no cabeçalho dos arquivos modificados/criados (exemplo: `Copyright (C) 2026 Seu Nome`).
3. Mantenha os créditos ao autor original e contribuidores.
