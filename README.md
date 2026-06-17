<div align="center">
  <h1>Kei · كي</h1>
  <p><strong>"When are you free?"</strong></p>
  <p>Self-hosted, open source team scheduling — works offline, supports Persian & English, syncs with Google Calendar and Outlook.</p>

  ![License](https://img.shields.io/badge/license-MIT-green)
  ![Status](https://img.shields.io/badge/status-Phase%201%20development-blue)
  ![Stack](https://img.shields.io/badge/stack-Angular%20%7C%20Go%20%7C%20PostgreSQL-purple)
</div>

---

## What is Kei?

Kei answers the question every distributed team asks dozens of times a week: **كي؟ — When are you free?**

- **Works offline** — runs on your own server, zero cloud dependency
- **Multi-timezone** — each team member sees times in their local timezone
- **Persian + English** — full RTL support, Jalali and Gregorian calendars
- **Calendar sync** — bidirectional sync with Google Calendar and Outlook *(Phase 3)*
- **Open source** — MIT license, self-host in one command

---

## Current Status

| Phase | Description | Status |
|---|---|---|
| Phase 1 | Availability board | 🔨 In development |
| Phase 2 | Meeting proposals | ⏳ Planned |
| Phase 3 | Calendar sync + intelligence | ⏳ Planned |
| Phase 4 | Multi-tenant + open source launch | ⏳ Planned |

---

## Quick Start (Self-Hosted)

**Requirements:** Docker + Docker Compose

```bash
# 1. Clone the repo
git clone https://github.com/your-handle/kei.git
cd kei

# 2. Configure environment
cp .env.example .env
# Edit .env — set your passwords and JWT secrets

# 3. Start
docker compose up -d

# 4. Open
# http://localhost:4200
# First visit runs the setup wizard
```

---

## Development Setup

**Requirements:** Go 1.23+, Node 20+, PostgreSQL 15+

```bash
# Backend
cd backend
cp ../.env.example ../.env
go mod download
go run cmd/server/main.go

# Frontend (separate terminal)
cd frontend
npm install
npm start
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Angular 19 + Signals |
| Backend | Go |
| Database | PostgreSQL 15+ |
| Real-time | WebSocket |
| Deployment | Docker Compose |

---

## Contributing

Kei is open source and welcomes contributions.  
Read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a PR.

---

## License

MIT — see [LICENSE](LICENSE)

---

<div align="center">
  <sub>Built for teams that can't always rely on the internet — but always need to meet.</sub>
</div>
