# CLAUDE.md — Kei (كي)

> Read this file at the start of every session. It is the single source of truth for all decisions made about this project.

---

## What is Kei?

Kei (كي — Persian for "when?") is a self-hosted, open source team scheduling tool. It gives distributed teams shared visibility into availability, an async meeting proposal flow, and bidirectional calendar sync — without depending on any cloud SaaS, subscription, or stable internet connection.

It runs entirely on a team's own server. Works offline. Supports English and Persian (RTL). Syncs with Google Calendar and Outlook.

**GitHub:** github.com/[your-handle]/kei  
**License:** MIT

---

## Current Phase: PHASE 1 — Availability Board

### Phase 1 scope — build ONLY this:
- User authentication (JWT, local — no external OAuth)
- First-time setup wizard
- Per-user timezone + language + calendar preference
- Weekly availability grid (free / busy / focus slots)
- Recurring availability templates
- Team dashboard — all members in one view
- Overlap window — core hours highlighted across timezones
- Real-time updates via WebSocket
- Role-based access: Admin / Member / Viewer
- English + Persian UI with full RTL layout
- Jalali + Gregorian calendar display

### Phase 1 does NOT include — do not build yet:
- Meeting proposals of any kind
- Calendar sync (Google, Outlook)
- Notifications (Slack, email)
- Analytics
- Redis
- Multi-tenant / multi-org
- Guest links

---

## Monorepo Structure

```
kei/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/
│   │   ├── availability/
│   │   ├── config/
│   │   ├── db/
│   │   ├── middleware/
│   │   ├── organization/
│   │   ├── realtime/
│   │   ├── server/
│   │   └── user/
│   ├── migrations/
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── core/
│   │   │   │   ├── auth/
│   │   │   │   ├── i18n/
│   │   │   │   └── services/
│   │   │   ├── features/
│   │   │   │   ├── auth/
│   │   │   │   ├── availability/
│   │   │   │   ├── dashboard/
│   │   │   │   └── settings/
│   │   │   └── shared/
│   │   │       ├── components/
│   │   │       └── pipes/
│   │   ├── assets/
│   │   │   └── i18n/
│   │   │       ├── en.json
│   │   │       └── fa.json
│   │   └── environments/
│   ├── angular.json
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.example
├── CLAUDE.md
├── PRODUCT.md
├── README.md
└── .gitignore
```

---

## Tech Stack

| Layer | Choice | Notes |
|---|---|---|
| Frontend | Angular 19 + Signals | `@angular/localize` for i18n, `ngx-translate` for runtime switching |
| Backend | Go 1.23+ | Standard library where possible, minimal dependencies |
| Database | PostgreSQL 15+ | `tstzrange`, exclusion constraints for conflict detection |
| Real-time | WebSocket | Native Go `gorilla/websocket` |
| Auth | JWT | `golang-jwt/jwt` — access token 15min, refresh 7 days |
| Jalali | date-fns-jalali | Frontend only — all DB values are Gregorian UTC |
| RTL | Angular CDK | `dir` attribute switching on root element |
| Containerization | Docker + Docker Compose | Single command self-host |

---

## Database Conventions

- **All timestamps:** `TIMESTAMPTZ` (UTC always). Never store local time.
- **All IDs:** `UUID` using `gen_random_uuid()`
- **Time ranges:** `TSTZRANGE` for availability slots and meetings
- **Conflict prevention:** PostgreSQL exclusion constraints on `availability_slots`
- **Migrations:** Sequential numbered files in `backend/migrations/` — e.g. `001_init.sql`, `002_add_users.sql`
- **No ORM:** Raw SQL with `pgx/v5` driver

---

## API Conventions

- Base path: `/api/v1/`
- Auth: Bearer JWT in Authorization header
- All times in requests/responses: ISO 8601 UTC (e.g. `2026-07-01T10:00:00Z`)
- Errors: `{ "error": "message", "code": "ERROR_CODE" }`
- Success lists: `{ "data": [...], "total": N }`
- WebSocket: `/ws/availability` — JSON messages with `type` field

### Phase 1 endpoints:
```
POST   /api/v1/auth/setup
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
DELETE /api/v1/auth/logout
POST   /api/v1/auth/password-reset/request
POST   /api/v1/auth/password-reset/confirm

GET    /api/v1/users
POST   /api/v1/users/invite
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
PUT    /api/v1/users/:id/role
DELETE /api/v1/users/:id

GET    /api/v1/availability/:user_id
PUT    /api/v1/availability/:user_id
GET    /api/v1/availability/:user_id/recurring
PUT    /api/v1/availability/:user_id/recurring
POST   /api/v1/availability/:user_id/import

GET    /api/v1/team/availability
GET    /api/v1/team/overlap

GET    /api/v1/settings
PUT    /api/v1/settings
GET    /api/v1/settings/blocked-days
POST   /api/v1/settings/blocked-days
DELETE /api/v1/settings/blocked-days/:id

WS     /ws/availability
```

---

## Core Data Model (Phase 1)

```sql
organizations (id, name, slug, timezone, overlap_start, overlap_end, created_at)

users (id, org_id, name, email, password[bcrypt], role, timezone, language, calendar_pref, is_active, created_at)

availability_slots (id, user_id, slot_range[TSTZRANGE], status[free|busy|focus], is_override, recurrence_id, created_at)
-- EXCLUDE USING GIST (user_id WITH =, slot_range WITH &&) WHERE status IN ('busy','focus')

recurrence_templates (id, user_id, pattern[daily|weekly], days_of_week[int[]], start_time, end_time, status, valid_from, valid_until)

blocked_days (id, org_id, blocked_date, reason, created_by, created_at)

audit_log (id, org_id, actor_id, action, target_type, target_id, metadata[JSONB], created_at)
```

---

## Go Code Conventions

- Package per domain: `internal/auth`, `internal/availability`, `internal/user`
- Handler → Service → Repository pattern
- No global state — dependency injection via structs
- Errors wrapped with context: `fmt.Errorf("availability.GetSlots: %w", err)`
- Config from environment variables only — no config files in production
- Tests in `_test.go` files alongside implementation

### Environment variables:
```
DATABASE_URL=postgres://kei:kei@localhost:5432/kei?sslmode=disable
JWT_SECRET=change-me-in-production
JWT_REFRESH_SECRET=change-me-in-production-too
PORT=8080
ENV=development
CORS_ORIGIN=http://localhost:4200
```

---

## Angular Conventions

- Standalone components (no NgModules)
- Signals for state management
- `inject()` function — no constructor injection
- Lazy-loaded feature routes
- i18n: `ngx-translate` for runtime EN/FA switching
- RTL: toggle `dir="rtl"` on `<html>` element when language = FA
- Jalali: `date-fns-jalali` for display only — all API calls use ISO UTC dates
- HTTP interceptor handles JWT attach + refresh

---

## Key Business Rules (Phase 1)

1. **Busy = hard block.** Cannot be overridden by anyone except admin.
2. **Focus = soft block.** Overridable in Phase 2 by organizers with explicit acknowledgment.
3. **All times stored UTC.** Display layer converts to user's timezone.
4. **Jalali is display only.** All DB values and API payloads use Gregorian UTC.
5. **Overlap window** stored as UTC time range on the organization. Each user's dashboard highlights it in their local time.
6. **Setup wizard** runs once — detected by empty users table. Never runs again.
7. **Admin cannot delete their own account** if they are the only admin.
8. **Adjacent same-type slots** are merged into a single range on save.

---

## Frontend Design Tokens (styles/tokens.scss)

All colours are CSS custom properties. **No hardcoded values outside tokens.scss.**

| Token | Value | Usage |
|---|---|---|
| `--kei-bg-base` | `#0D1117` | App background, sidebar |
| `--kei-bg-surface` | `#161B22` | Topbar, cards |
| `--kei-bg-overlay` | `#1C2330` | Modals, dropdowns |
| `--kei-bg-card` | `#21293A` | Elevated card surfaces |
| `--kei-accent` | `#58A6FF` | Primary accent (GitHub blue) |
| `--kei-accent-glow` | `rgba(88,166,255,0.12)` | Active/hover backgrounds |
| `--kei-accent-border` | `rgba(88,166,255,0.25)` | Active/focus borders |
| `--kei-border` | `#2A3441` | Default borders |
| `--kei-border-hover` | `#334155` | Hover / emphasis borders |
| `--kei-text-primary` | `#E6EDF3` | Primary text |
| `--kei-text-secondary` | `#8B949E` | Secondary text |
| `--kei-text-muted` | `#4A5568` | Disabled / muted text |
| `--kei-free-bg/text` | `#1A3A2A / #2EA043` | Free slot |
| `--kei-busy-bg/text` | `#3A1A1A / #F85149` | Busy slot |
| `--kei-focus-bg/text` | `#1A1F3A / #58A6FF` | Focus slot |
| `--kei-overlap-bg/text` | `#0D1A2E / #58A6FF` | Overlap highlight |

## Frontend Component Architecture (Step 9+)

```
AppComponent
  └── AppShellComponent          ← full-height flex layout
        ├── SidebarComponent     ← 48px, #0D1117 bg, icon nav + "كي" logo
        ├── TopbarComponent      ← 44px, #161B22 bg, [title]/[subtitle] inputs
        └── MainContentComponent ← flex:1, scrollable, <router-outlet> inside
```

Services:
- `KeiThemeService` — `currentLocale$` BehaviorSubject, `toggleLocale()`, applies `dir` + `lang` + ngx-translate
- `KeiIconComponent` — renders Tabler Icons (outline) as inline SVG by name

i18n:
- Runtime switching: ngx-translate (`public/i18n/en.json`, `public/i18n/fa.json`)
- Compile-time locale data: `@angular/localize` with `src/locale/messages.fa.xlf`
- Locale IDs: `en-US` (source) and `fa-IR`

## What to Build Next (Phase 1 sequence)

Build in this order — each step is independently testable:

1. ✅ Docker Compose + PostgreSQL + migrations
2. ✅ Go server bootstrap (config, db connection, router)
3. ✅ Auth: setup wizard, login, JWT, refresh, logout, password reset
4. ✅ User management: invite, list, update, role change
5. ✅ Availability: set slots, get slots, merge adjacent, recurring templates
6. ✅ Team dashboard API: all members' availability + overlap calculation
7. ✅ WebSocket: real-time availability broadcast
8. ✅ Settings: working hours, blocked days
9. ✅ Angular: project scaffold + i18n + RTL setup
10. ✅ Angular: auth pages (login, setup wizard)
11. Angular: availability grid component
12. Angular: team dashboard
13. Angular: settings pages
14. Angular: WebSocket integration

---

## Session Startup Checklist

At the start of every Claude Code session:
1. Read this file
2. Check current git status (`git status`)
3. Ask the developer: "What are we building this session?"
4. Build only what's in Phase 1 scope
5. After each feature: remind the developer to test before moving on

---

## Auth Flow (Step 10)

- **Cookie-based** — access token in httpOnly cookie, set by backend on login/setup. Frontend stores nothing in localStorage.
- **Session rehydration** — `APP_INITIALIZER` calls `GET /api/v1/users/me`; if 401, session is cleared and guards redirect to `/login`.
- **Guards:** `authGuard` (protect shell routes), `noAuthGuard` (prevent logged-in users hitting `/login`), `setupGuard` (check `GET /api/v1/auth/setup/status`; redirect to `/login` if setup done).
- **Interceptor** — `authInterceptor` adds `withCredentials: true` to every request; on 401 from non-auth endpoints, clears session and navigates to `/login`.
- **Form overrides** — `src/styles/_forms.scss` remaps Angular Material MDC CSS variables to dark tokens. No `!important` except on the text-field background.

*Last updated: Step 10 complete — auth pages, AuthService, guards, interceptor, form styles*  
*Next: Step 11 — Angular availability grid component*  
*Next phase: Phase 2 — Meeting Proposals (do not start until Phase 1 is fully tested)*
