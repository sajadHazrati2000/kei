# Kei (كي) — Product Requirements Document

**Version:** 1.0  
**Status:** Living document — update as product evolves  
**Repository:** github.com/[your-handle]/kei  
**License:** MIT  

---

## 1. Vision

Kei answers the question every distributed team asks dozens of times a week: **"كي؟ — When are you free?"**

A self-hosted, open source scheduling tool that works offline, respects timezones, speaks Persian and English, and syncs with the calendars your team already uses — without depending on any cloud service, subscription, or stable internet connection.

---

## 2. Problem Statement

Teams with members across regions waste significant time coordinating meetings because:

- No shared visibility into who is actually free vs busy vs in deep work
- Timezone math is done manually — error-prone and slow
- Meeting proposals require multiple Slack messages before a time is agreed
- Existing SaaS tools (Reclaim, Clockwise) are cloud-dependent, potentially blocked in some regions, and require ongoing subscriptions
- No tool handles Jalali + Gregorian calendars natively alongside bidirectional external calendar sync

---

## 3. Personas

### P1 — Organizer
Any team member who initiates meetings. Needs to propose a time, see who's free, get confirmation fast, and have the meeting appear in everyone's external calendar automatically.

### P2 — Attendee
Any team member invited to a meeting. Needs to see what's being asked, respond quickly from any device, and have conflicts surfaced before they accept.

### P3 — Admin
Team or IT administrator. Needs to configure the system, manage users, set org-wide defaults, and keep the self-hosted instance running.

### P4 — Guest
External participant with no account. Receives a time-limited link, sees minimal meeting details, and can accept or suggest an alternative time without logging in.

### P5 — Viewer
Stakeholder with read-only access. Can see team availability and meeting schedules but cannot create or respond to proposals.

---

## 4. Core Principles

1. **Offline first** — every core feature works on a local network with no internet
2. **UTC everywhere** — all times stored as UTC, displayed in each user's local timezone
3. **Trust through transparency** — conflicts are never silently resolved; the user always decides
4. **Scope discipline** — this is a scheduling tool, not a project manager or task tracker
5. **Open by default** — MIT license, self-hostable in one command, no telemetry without opt-in

---

## 5. Scope Boundaries

### In scope
- Team availability management (free / busy / focus)
- Meeting proposals: async, organizer-pick, poll, counter-propose
- Conflict detection: hard (busy) and soft (focus)
- Two-way calendar sync: Google Calendar + Outlook / Microsoft 365
- Multi-timezone support with configurable team overlap window
- Jalali (Shamsi) + Gregorian dual calendar display
- English + Persian (Farsi) UI with full RTL support
- Notifications: in-app, Slack webhook, email (SMTP optional)
- Guest participation via time-limited links
- Self-hosted (Docker) + cloud deployment
- Role-based access: Admin / Member / Viewer

### Out of scope (explicit — do not add)
- Task management
- Project tracking
- Video/audio calling
- File sharing
- Time tracking / invoicing
- Native mobile apps (web only)

---

## 6. Success Metrics

### Phase 1
- Team uses the availability grid daily without prompting
- Zero "what time are you free?" Slack messages within 2 weeks of adoption
- Conflict detection catches 100% of hard double-bookings

### Phase 2
- Meeting proposals confirmed in under 24 hours on average
- Less than 2 rounds of back-and-forth per meeting
- Organizers report less friction than current process

### Phase 3
- Calendar sync runs without manual intervention for 30 consecutive days
- Slot suggestions accepted (not manually overridden) >60% of the time
- Zero unresolved timezone errors in meeting times

### Phase 4
- Successfully deployed by at least 3 external teams via Docker
- GitHub stars >100 within 3 months of public release
- At least one community-contributed language pack

---

## 7. Technical Constraints

| Constraint | Decision |
|---|---|
| Frontend | Angular 19 + Signals |
| Backend | Go |
| Database | PostgreSQL 15+ (`tstzrange`, exclusion constraints) |
| Cache / locking | Redis (Phase 3+) |
| Real-time | WebSocket (native Go) |
| Deployment | Docker Compose (self-hosted) + standard cloud (Phase 4) |
| Calendar sync | OAuth2 — Google Calendar API + Microsoft Graph API |
| i18n | Angular `@angular/localize` + `ngx-translate` for runtime switching |
| Jalali calendar | `date-fns-jalali` or `jalali-moment` |
| Auth | JWT (access + refresh tokens) — no external OAuth dependency for login |
| Timezone | All DB timestamps `TIMESTAMPTZ` (UTC). Per-user `timezone` field (IANA format) |

---

## 8. Architecture Overview

```
┌─────────────────────────────────────────────┐
│                  Angular 19                  │
│   RTL layout · Jalali/Gregorian · i18n       │
│   WebSocket client · Calendar sync UI        │
└─────────────────┬───────────────────────────┘
                  │ HTTP + WebSocket
┌─────────────────▼───────────────────────────┐
│                  Go API server               │
│   REST endpoints · WebSocket hub             │
│   Sync engine (Google + Outlook OAuth2)      │
│   Scheduler jobs · Webhook delivery          │
│   Timezone conversion layer                  │
└──────────┬──────────────────┬───────────────┘
           │                  │
┌──────────▼──────┐  ┌────────▼──────────────┐
│  PostgreSQL 15+ │  │  Redis (Phase 3+)      │
│  UTC timestamps │  │  Slot locking          │
│  tstzrange      │  │  Job queue             │
│  Audit log      │  │  Notification retry    │
└─────────────────┘  └────────────────────────┘
```

---

---

# PHASE 1 — Availability Board

## PRD: Phase 1

### Goal
Give the team shared, real-time visibility into who is free, busy, or in focus mode — across timezones — without any meeting proposal flow. This alone should eliminate the majority of "when are you free?" messages.

### Deliverables
1. User authentication (JWT, local — no external OAuth)
2. First-time setup wizard
3. Per-user timezone configuration
4. Weekly availability grid (free / busy / focus slots)
5. Recurring availability templates
6. Team dashboard — all members' availability in one view
7. Overlap window — configurable core hours highlighted across all timezones
8. Real-time updates via WebSocket
9. Role-based access (Admin / Member / Viewer)
10. English + Persian UI with RTL layout
11. Jalali + Gregorian calendar display toggle

### What Phase 1 does NOT include
- Meeting proposals
- Calendar sync
- Notifications (beyond in-app)
- Analytics

### Phase 1 Definition of Done
- Any team member can open the dashboard and immediately see who is free right now
- Availability changes reflect for all users within 1 second on local network
- Works with zero internet — fully functional on local network only
- UI renders correctly in both English (LTR) and Persian (RTL)
- Dates display correctly in both Jalali and Gregorian

---

## Scenarios: Phase 1

### SC-101 — First-time setup
**Actor:** Admin  
**Precondition:** Fresh Docker deployment  
**Flow:**
1. Admin navigates to app URL
2. Setup wizard detects empty database — runs automatically
3. Admin enters: organization name, their name, email, password, default timezone, preferred language (EN/FA)
4. Admin account created with full permissions
5. Default working hours set: 09:00–18:00, Mon–Fri (configurable)
6. Default overlap window set: 10:00–12:00 in org timezone
7. Dashboard loads

**Expected:** Completes in under 2 minutes. Wizard never runs again after first account exists.  
**Failure:** Database unreachable → "Cannot connect to database. Check your Docker setup."

---

### SC-102 — User login
**Actor:** Any user  
**Flow:**
1. User opens app, sees login screen
2. Language auto-detected from browser (EN or FA); user can switch
3. Enters email + password
4. JWT access token (15 min TTL) + refresh token (7 days) issued, stored in httpOnly cookies
5. Dashboard loads in user's configured language and timezone

**Expected:** Under 1 second on local network  
**Wrong credentials:** "Incorrect email or password" — no hint which field is wrong  
**Inactive account:** "Your account has been deactivated. Contact your admin."

---

### SC-103 — Login with internet cut
**Actor:** Any user  
**Precondition:** Internet completely unavailable. Server running on local network.  
**Flow:**
1. User opens app via local IP or hostname
2. Authentication runs on local server — no external dependency
3. Full app loads and functions normally

**Expected:** Zero degradation. This is a core product guarantee, not a fallback.

---

### SC-104 — Session refresh
**Actor:** Any user  
**Flow:**
1. Access token expires (15 min)
2. App silently requests new access token using refresh token
3. If refresh succeeds: user never notices
4. If refresh token also expired: redirect to login
5. Before redirect: current form state saved to sessionStorage
6. After login: user returned to last page with state restored

---

### SC-105 — Admin invites a team member
**Actor:** Admin  
**Flow:**
1. Admin opens Settings → Users → Invite
2. Enters name, email, role (Member / Viewer), timezone, preferred language
3. Account created with temporary password
4. If SMTP configured: welcome email sent with login link
5. If no SMTP: admin copies the invite link and shares manually (works offline)
6. New member logs in, prompted to change password on first login

**Expected:** Invite works entirely on local network without email if needed.

---

### SC-106 — User configures their timezone and calendar
**Actor:** Member  
**Flow:**
1. Member opens Profile Settings
2. Selects timezone from IANA list (e.g. Asia/Tehran, Europe/Berlin)
3. Selects preferred calendar display: Jalali or Gregorian
4. Selects preferred language: English or Persian
5. Saves — UI immediately re-renders in new language/calendar/timezone
6. All times on the dashboard now show in their local timezone

**Expected:** Switching from Gregorian to Jalali instantly re-renders all dates. No page reload required.

---

### SC-107 — Member sets weekly availability
**Actor:** Member  
**Flow:**
1. Member opens Availability page
2. Sees 7-day grid with 30-minute slots, times shown in their local timezone
3. Dates shown in their preferred calendar (Jalali or Gregorian)
4. Clicks slots to cycle: Free (green) → Busy (red) → Focus (amber) → Free
5. Saves
6. Stored as UTC time ranges in PostgreSQL
7. Team dashboard updates for all members in real-time via WebSocket

**Slot states:**
- **Free** — available for meetings
- **Busy** — unavailable, hard block, reason private
- **Focus** — deep work, soft block, overridable by organizers

**Expected:** Changes visible to all members within 1 second on local network  
**Edge case:** Adjacent same-type slots → merged into a single range automatically

---

### SC-108 — Member sets recurring availability template
**Actor:** Member  
**Flow:**
1. Member opens Availability → Recurring Settings
2. Defines weekly pattern: e.g. Sun–Thu 09:00–17:00 Free, Wed 10:00–13:00 Focus
3. System generates slots for the next 4 weeks (configurable window)
4. Individual days can be overridden without breaking the template
5. Overrides shown with a distinct indicator on the grid

**Expected:** Template and overrides stored separately. Re-generating template does not wipe overrides.

---

### SC-109 — View team availability dashboard
**Actor:** Any user  
**Flow:**
1. User opens Team Dashboard
2. Sees all active members on a single weekly grid
3. Each member's row shows their availability in **the viewer's local timezone**
4. Overlap window highlighted across all rows — shows when everyone (or most people) are free
5. Slot color per member:
   - Green — free
   - Red — busy
   - Amber — focus
   - Gray — not set
6. Hovering a slot: tooltip shows member name, status, their local time equivalent
7. Members sorted by: most available this week (default), or alphabetically

**Expected:** Loads within 1 second for up to 50 members  
**Edge case:** Member in different timezone — their row clearly labels their local time alongside the viewer's time

---

### SC-110 — Admin configures overlap window
**Actor:** Admin  
**Flow:**
1. Admin opens Settings → Team Hours
2. Sets default working days and hours per org timezone
3. Sets overlap window: the core hours when all team members are expected online
4. Overlap window stored as UTC range
5. Each member's dashboard highlights this window in their local time
6. Example: overlap set to 10:00–12:00 Tehran (UTC+3:30) → Berlin member sees 07:30–09:30 highlighted

**Expected:** Overlap window recalculates automatically if a member changes their timezone  
**Edge case:** No common overlap exists (e.g. Tehran + Vancouver) → system shows "No full-team overlap this week" with the best partial overlap highlighted

---

### SC-111 — Real-time availability update
**Actor:** System  
**Trigger:** Any member saves an availability change  
**Flow:**
1. Member saves updated slots
2. Server broadcasts change via WebSocket to all connected clients
3. Team dashboard updates without page refresh
4. If a client's WebSocket is disconnected: change received on next reconnect via missed-events catch-up

**Expected:** Update visible within 1 second on local network  
**Fallback:** If WebSocket unavailable: clients poll REST endpoint every 30 seconds

---

### SC-112 — Admin sets no-meeting days
**Actor:** Admin  
**Flow:**
1. Admin opens Settings → Blocked Days
2. Adds specific dates: national holidays, company off-days
3. Days marked with distinct styling on team grid
4. All members' availability on those days shown as blocked
5. In Phase 2: meeting proposals on blocked days rejected at submission

---

### SC-113 — RTL layout switch
**Actor:** Persian-language user  
**Flow:**
1. User selects Persian (فارسی) as their language
2. Entire UI mirrors to RTL layout
3. Navigation, grid, forms, tooltips, dropdowns — all RTL
4. Jalali calendar automatically activated as default
5. All Persian text uses correct typographic rules
6. User can switch back to English (LTR) at any time

**Expected:** RTL and LTR render correctly simultaneously for different users viewing the same dashboard

---

### SC-114 — Role-based access enforcement
**Actor:** Viewer  
**Flow:**
1. Viewer logs in
2. Can see team availability dashboard
3. Cannot see availability edit controls
4. Any attempt to call write endpoints returns 403
5. Navigation hides actions unavailable to their role

| Action | Admin | Member | Viewer |
|---|---|---|---|
| View team grid | ✅ | ✅ | ✅ |
| Set own availability | ✅ | ✅ | ❌ |
| Invite users | ✅ | ❌ | ❌ |
| Edit system settings | ✅ | ❌ | ❌ |
| Override others' availability | ✅ | ❌ | ❌ |

---

### SC-115 — Admin overrides a member's availability
**Actor:** Admin  
**Precondition:** Member is unavailable or unreachable  
**Flow:**
1. Admin opens a member's availability page
2. Edits slots directly
3. Override logged in audit trail: actor, timestamp, previous value, new value
4. Member notified via in-app notification on next login

---

### Phase 1 — Edge Cases

| Scenario | Expected Behavior |
|---|---|
| Member hasn't set any availability | Shown as all-gray on team grid with "Not set" label |
| Two members save availability simultaneously | PostgreSQL row-level locking — both succeed independently, no conflict |
| Member sets 100% busy for the week | Respected — shown as fully red. No system warning. |
| Timezone not set on user account | Defaults to org timezone. Prompted to set on first login. |
| Jalali date in a year with different month lengths | `date-fns-jalali` handles this — no custom logic needed |
| Admin deletes their own account | Blocked: "Assign another admin first" |
| Browser in RTL system but user picks English | LTR renders correctly regardless of browser locale |
| Member changes timezone mid-week | All existing slots recalculate display instantly. UTC values unchanged. |
| WebSocket connection drops | Client retries: 1s → 2s → 4s → 8s → 30s max. Falls back to polling. |
| 50+ members on the team grid | Virtual scrolling — only visible rows rendered |

---

---

# PHASE 2 — Meeting Proposals

## PRD: Phase 2

### Goal
Give the team a structured, async-first way to propose, respond to, and confirm meetings — eliminating the back-and-forth Slack thread entirely.

### Deliverables
1. Three proposal flows: async (propose → respond), organizer-pick (notify only), poll (multi-slot vote)
2. Counter-propose flow
3. Recurring meetings
4. Cancel and reschedule
5. Hard conflict detection (busy blocks)
6. Soft conflict warning (focus blocks)
7. Near-conflict warning (back-to-back)
8. In-app notification center
9. Slack webhook notifications
10. Email notifications (SMTP optional)
11. Per-user notification preferences per event type
12. Guest participation via time-limited link
13. Meeting history and audit trail

### Phase 2 Definition of Done
- A meeting can be proposed, confirmed, and appear in all attendees' in-app calendars without a single Slack message
- Hard conflicts (busy) are never silently booked
- Focus conflicts require explicit override
- Notifications delivered within 5 seconds on local network
- Guest link works without an account

---

## Scenarios: Phase 2

### SC-201 — Propose a meeting (async flow)
**Actor:** Organizer  
**Flow:**
1. Organizer clicks New Meeting
2. Selects flow: Async (attendees confirm a time)
3. Fills in: title, required attendees, optional attendees, duration, notes
4. System queries availability of all required attendees
5. Shows top 3 suggested slots (scored — see SC-601 for full scoring logic)
6. Organizer selects a slot or picks a custom time
7. Proposal sent — status: `pending`
8. All attendees notified per their notification preferences

**Expected:** Suggestions returned within 500ms for teams up to 50 members  
**Edge case:** No conflict-free slot in 14 days → least-conflict options shown with explicit conflict summary

---

### SC-202 — Propose a meeting (organizer-pick flow)
**Actor:** Organizer  
**Flow:**
1. Organizer selects flow: Organizer picks time
2. Picks a time directly — no attendee input requested
3. Proposal sent immediately with status: `confirmed`
4. Attendees notified: "You have been added to [Meeting] on [time]"
5. Slot locked in all attendees' availability

**Expected:** Fastest flow — zero round-trips. Used for mandatory meetings.  
**Conflict behavior:** If slot conflicts with a busy block → hard block warning shown to organizer before confirming

---

### SC-203 — Propose a meeting (poll flow)
**Actor:** Organizer  
**Flow:**
1. Organizer selects flow: Poll — let attendees vote
2. Proposes 2–5 time options
3. Poll sent to all attendees
4. Each attendee votes: Yes / If needed / No for each option
5. Results shown in real-time to organizer
6. Organizer reviews votes and confirms one slot
7. Meeting confirmed — attendees notified

**Expected:** Poll results update in real-time via WebSocket  
**Deadline:** Organizer can set a poll expiry (default: 48 hours). After expiry: organizer prompted to pick based on current votes.

---

### SC-204 — Accept a proposal
**Actor:** Attendee  
**Flow:**
1. Attendee receives notification
2. Opens proposal — sees title, time in their local timezone and calendar (Jalali/Gregorian), organizer, attendees, notes
3. Clicks Accept
4. Response: `accepted`. Slot: `tentative`
5. Organizer sees real-time update
6. When all required attendees accept → status: `confirmed`. Slot: hard-locked.
7. All attendees notified of confirmation

---

### SC-205 — Decline a proposal
**Actor:** Attendee  
**Flow:**
1. Attendee clicks Decline
2. Optional: adds reason (visible to organizer only, never to other attendees)
3. If required attendee declines:
   - Status → `needs_rescheduling`
   - Organizer notified with reason
   - Organizer prompted: reschedule or remove attendee
4. If optional attendee declines:
   - Meeting proceeds
   - Attendee marked as not attending
   - Organizer quietly notified

---

### SC-206 — Counter-propose a time
**Actor:** Attendee  
**Flow:**
1. Attendee clicks Suggest Another Time
2. Team availability grid shown with attendee's free slots highlighted
3. Attendee picks alternative slot
4. Adds optional note
5. Counter sent to organizer
6. Organizer accepts counter or initiates another round
7. Maximum 2 rounds before system flags for manual coordination
8. All counter-proposals form a visible thread on the meeting record

---

### SC-207 — Hard conflict detection
**Actor:** System  
**Trigger:** Proposal submitted for a slot where a required attendee has a busy block or confirmed meeting  
**Flow:**
1. Server checks `tstzrange` overlap against busy slots and confirmed meetings
2. Overlap found → `409 Conflict` returned
3. Frontend: "This time conflicts with [Name]'s schedule. Choose another slot."
4. Proposal not saved

**Expected:** Blocked at database level via exclusion constraint — not just application logic

---

### SC-208 — Soft conflict warning (focus override)
**Actor:** Organizer  
**Trigger:** Proposed time overlaps a required attendee's focus block  
**Flow:**
1. System detects focus overlap — not a hard block
2. Warning shown: "[Name] has focus time at this slot"
3. Organizer must explicitly check "Override focus block" to proceed
4. If overridden: affected member receives priority-flagged notification

---

### SC-209 — Near-conflict warning
**Actor:** System  
**Trigger:** Proposed meeting leaves less than 15 minutes between meetings for any attendee  
**Flow:**
1. Gap detected after slot selection
2. Inline warning: "[Name] has a meeting ending at [time] — only [N] minutes before this one"
3. Not a blocker — organizer can proceed

**Expected:** Gap threshold configurable by admin (default: 15 minutes)

---

### SC-210 — Cancel a confirmed meeting
**Actor:** Organizer  
**Flow:**
1. Organizer clicks Cancel on a confirmed meeting
2. Adds optional cancellation reason
3. All attendees notified
4. Slot released atomically in all attendees' availability
5. Meeting record retained with status `cancelled`

**Expected:** Atomic — no attendee left with a blocked slot after cancellation  
**Urgent cancellation:** If cancelled within 30 minutes of start → notification marked urgent

---

### SC-211 — Reschedule a confirmed meeting
**Actor:** Organizer  
**Flow:**
1. Organizer clicks Reschedule
2. System re-queries current availability of all attendees
3. New slot suggestions shown
4. Organizer picks new slot
5. Old slot released, new slot tentatively locked
6. Attendees re-notified — must re-confirm
7. Reschedule history logged on meeting record

---

### SC-212 — Recurring meeting
**Actor:** Organizer  
**Flow:**
1. Organizer enables recurrence: daily / weekly / biweekly / monthly
2. Sets end date or occurrence count
3. Instances generated and linked to recurrence group
4. Each instance is an independent record
5. Edit options: this instance / all future instances / entire series
6. Cancel one instance → series unaffected

**Edge case:** Instance falls on a blocked day → auto-cancelled, organizer notified

---

### SC-213 — Notification preferences
**Actor:** Member  
**Flow:**
1. Member opens Profile → Notifications
2. For each event type, selects delivery channel:
   - In-app (always on, cannot disable)
   - Slack webhook (if configured by admin)
   - Email (if SMTP configured)
3. Settings saved per user

| Event | Default channels |
|---|---|
| New proposal | In-app + Slack |
| Accepted | In-app |
| Declined | In-app + Slack |
| Meeting confirmed | In-app + Slack |
| Meeting cancelled | In-app + Slack + Email |
| Reminder (15 min) | In-app + Slack |
| Focus block overridden | In-app + Slack |

---

### SC-214 — Guest participation
**Actor:** Organizer  
**Flow:**
1. Organizer adds an external participant (email, no account)
2. System generates time-limited public link (expires 48 hours)
3. Organizer shares link manually
4. Guest opens link — no login required
5. Sees: meeting title, proposed time in their detected timezone, organizer name only
6. Guest accepts or suggests alternative time
7. Response recorded. Organizer notified.

**Security:** Guest link exposes only: title, time, organizer name. No team data, no other attendees visible.

---

### SC-215 — Concurrent proposal race condition
**Actor:** Two organizers proposing overlapping meetings simultaneously  
**Flow (Phase 2 — PostgreSQL row-level lock):**
1. Both proposals submitted within milliseconds
2. PostgreSQL serializes via row-level locking
3. First writer wins
4. Second receives `409 Conflict`
5. Second organizer prompted to choose different slot

---

### Phase 2 — Edge Cases

| Scenario | Expected Behavior |
|---|---|
| Proposal with no required attendees | Blocked: "Add at least one required attendee" |
| Proposal for a past time | Blocked: "Cannot schedule meetings in the past" |
| Duration set to 0 | Blocked: minimum 15 minutes |
| Attendee's account deleted mid-proposal | Proposal → `needs_review`. Organizer notified. |
| DB connection lost mid-proposal | Transaction rolled back fully. No partial state. |
| Organizer double-clicks submit | Idempotency key on proposal endpoint — second request returns existing proposal |
| Poll with no votes after expiry | Organizer notified to pick manually |
| Meeting title > 200 characters | Blocked at frontend + backend |
| Attendee in different timezone confirms | Meeting time shown correctly in their timezone throughout |
| Guest link expired | "This link has expired. Contact the meeting organizer." |
| All attendees decline | Meeting → `cancelled`. Organizer notified. |

---

---

# PHASE 3 — Calendar Sync + Intelligence

## PRD: Phase 3

### Goal
Connect Kei to the calendars team members already use, make conflict detection automatic, and let the system suggest optimal meeting times instead of requiring manual slot-hunting.

### Deliverables
1. Google Calendar OAuth2 integration — bidirectional sync
2. Outlook / Microsoft 365 OAuth2 integration — bidirectional sync
3. User-configurable sync interval
4. Manual sync trigger
5. Conflict resolution UI — manual, user always decides
6. Smart slot suggestion engine with scoring
7. Auto-conflict detection when availability changes
8. Meeting load analytics
9. Focus time suggestions
10. Redis-based slot locking for race conditions
11. Sync status dashboard per user

### Phase 3 Definition of Done
- A meeting confirmed in Kei appears in Google Calendar / Outlook within one sync cycle
- An event added in Google Calendar / Outlook appears in Kei's availability within one sync cycle
- Conflicts between Kei and external calendars are never silently resolved
- Slot suggestions accepted without manual override >60% of the time

---

## Scenarios: Phase 3

### SC-301 — Connect Google Calendar
**Actor:** Member  
**Flow:**
1. Member opens Profile → Calendar Sync
2. Clicks Connect Google Calendar
3. OAuth2 consent screen shown (Google)
4. Member grants access: read events + write events
5. OAuth tokens stored encrypted in PostgreSQL
6. Initial sync runs: external events pulled into Kei as busy/free slots
7. Sync status shown: "Last synced: just now"

**Expected:** Works in cloud deployment. In self-hosted: requires admin to configure Google OAuth credentials in system settings first (Client ID + Secret from Google Cloud Console).  
**Offline:** Sync paused when internet unavailable. Resumes automatically when connectivity restores.

---

### SC-302 — Connect Outlook / Microsoft 365
**Actor:** Member  
**Flow:** Identical to SC-301 using Microsoft Graph API OAuth2 flow.  
**Scope requested:** `Calendars.ReadWrite`  
**Expected:** Works for both personal Microsoft accounts and organizational M365 accounts.

---

### SC-303 — Auto-sync runs
**Actor:** System  
**Trigger:** Configured sync interval elapses (user-configurable: 5 / 15 / 30 / 60 minutes)  
**Flow:**
1. Go scheduler job fires for each user with active sync
2. Fetches events from external calendar since last sync timestamp
3. Compares with Kei's current availability and confirmed meetings
4. Three outcomes per event:
   - **New external event, no Kei conflict** → imported as busy slot automatically
   - **New external event, conflicts with Kei meeting** → conflict queued for user resolution
   - **Kei meeting confirmed since last sync** → pushed to external calendar as new event
5. Sync timestamp updated
6. Sync log entry written

**Expected:** Runs silently in background. User only notified when a conflict requires resolution.

---

### SC-304 — Manual sync trigger
**Actor:** Member  
**Flow:**
1. Member opens Calendar Sync settings
2. Clicks Sync Now
3. Sync runs immediately for their account
4. Results shown: "3 events imported, 1 meeting pushed, 0 conflicts"

---

### SC-305 — Conflict resolution (external event vs Kei meeting)
**Actor:** Member  
**Trigger:** Sync detects an external event overlapping a Kei confirmed meeting  
**Flow:**
1. Member notified: "Sync conflict detected — [External Event] overlaps [Kei Meeting]"
2. Member opens conflict resolution UI
3. Sees both events side by side with full details
4. Chooses:
   - Keep Kei meeting (external event treated as a duplicate — ignored)
   - Keep external event (Kei meeting → `needs_rescheduling`, organizer notified)
   - Both exist (acknowledge the overlap, take no action)
5. Decision logged in audit trail

**Expected:** System never auto-resolves. User always decides. Maximum 1 pending conflict per user shown at a time.

---

### SC-306 — Smart slot suggestion engine
**Actor:** System  
**Trigger:** Organizer opens new meeting form with attendees + duration  
**Scoring per candidate slot (next 14 days):**

| Criterion | Points |
|---|---|
| All required attendees free (no busy/focus) | 40 |
| Slot falls within team overlap window | 25 |
| ≥ 15 min gap from adjacent meetings for all attendees | 20 |
| Slot within each attendee's configured working hours | 15 |

**Flow:**
1. All candidate slots scored
2. Top 3 returned with human-readable reasoning:
   - "Best — everyone free, within overlap window"
   - "Good — one focus block (overridable)"
   - "Alternative — outside overlap window for [Name]"
3. Organizer selects or picks custom time

**Expected:** Results within 500ms for 50 members, 14-day window

---

### SC-307 — Meeting load analytics
**Actor:** Admin or Member  
**Flow:**
1. User opens Analytics
2. Personal view: meeting hours this week, count, focus time remaining, busiest day, trend vs last week
3. Admin view: team heatmap — meeting load per person
   - Green: < 3 hours/day
   - Amber: 3–5 hours/day
   - Red: > 5 hours/day
4. Drill down to individual breakdown
5. Export as CSV

---

### SC-308 — Redis slot locking (race condition prevention)
**Actor:** System  
**Trigger:** Two organizers propose meetings for overlapping slots simultaneously  
**Flow:**
1. Organizer A submits proposal for slot X
2. Redis lock acquired: key = `slot:{user_id}:{slot_range}`, TTL = 30 seconds
3. Organizer B submits proposal for overlapping slot X simultaneously
4. Redis rejects lock acquisition
5. Organizer B receives: "This slot was just taken. Please choose another."
6. On Organizer A's meeting confirmed: lock released, PostgreSQL record committed
7. On TTL expiry without confirmation: lock released, slot available again

---

---

# PHASE 4 — Multi-Tenant + Open

## PRD: Phase 4

### Goal
Make Kei deployable by any team in the world — one Docker command, zero configuration complexity — and open it fully to the community.

### Deliverables
1. Multi-organization support (schema-per-tenant isolation)
2. Self-host installer (single Docker Compose command)
3. Cloud deployment configuration
4. GitHub repository polished for open source: README, CONTRIBUTING, LICENSE, issue templates
5. External contributor documentation
6. Admin super-panel for managing multiple orgs (cloud only)
7. Org-level branding (logo, name)

### Phase 4 Definition of Done
- Any technical team can deploy Kei in under 5 minutes from the GitHub README
- Data from one organization is never queryable from another
- At least one external team has deployed and reported no critical issues

---

## Scenarios: Phase 4

### SC-401 — Self-host deployment
**Actor:** Technical admin of a new organization  
**Flow:**
1. Admin clones repo or downloads `docker-compose.yml`
2. Sets environment variables: `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `APP_URL`, `SECRET_KEY`
3. Runs: `docker compose up -d`
4. App available within 60 seconds
5. Setup wizard runs on first visit (SC-101)

**Expected:** Works on any Linux machine with Docker installed. No internet required after image pull.

---

### SC-402 — Organization isolation
**Actor:** System  
**Guarantee:** Every organization's data lives in its own PostgreSQL schema.  
**Expected:** A query in org A's context cannot return data from org B under any circumstance.

---

### SC-403 — Community contribution
**Actor:** External developer  
**Flow:**
1. Developer forks repo on GitHub
2. Reads CONTRIBUTING.md
3. Opens issue or picks existing one
4. Submits pull request
5. Maintainer reviews and merges

**Expected:** CONTRIBUTING.md covers: local dev setup, running tests, code style, PR checklist.

---

---

# Appendix A — Full Data Model

```sql
-- Organizations
CREATE TABLE organizations (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name         TEXT NOT NULL,
  slug         TEXT UNIQUE NOT NULL,
  timezone     TEXT NOT NULL DEFAULT 'UTC',
  overlap_start TIME,
  overlap_end   TIME,
  created_at   TIMESTAMPTZ DEFAULT now()
);

-- Users
CREATE TABLE users (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id       UUID REFERENCES organizations(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  email        TEXT UNIQUE NOT NULL,
  password     TEXT NOT NULL,
  role         TEXT CHECK (role IN ('admin','member','viewer')) DEFAULT 'member',
  timezone     TEXT NOT NULL DEFAULT 'UTC',
  language     TEXT CHECK (language IN ('en','fa')) DEFAULT 'en',
  calendar_pref TEXT CHECK (calendar_pref IN ('gregorian','jalali')) DEFAULT 'gregorian',
  is_active    BOOLEAN DEFAULT TRUE,
  created_at   TIMESTAMPTZ DEFAULT now()
);

-- Availability slots
CREATE TABLE availability_slots (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id        UUID REFERENCES users(id) ON DELETE CASCADE,
  slot_range     TSTZRANGE NOT NULL,
  status         TEXT CHECK (status IN ('free','busy','focus')) NOT NULL,
  is_override    BOOLEAN DEFAULT FALSE,
  recurrence_id  UUID REFERENCES recurrence_templates(id),
  source         TEXT CHECK (source IN ('manual','google','outlook','system')) DEFAULT 'manual',
  external_id    TEXT,
  created_at     TIMESTAMPTZ DEFAULT now(),
  EXCLUDE USING GIST (user_id WITH =, slot_range WITH &&)
    WHERE (status IN ('busy','focus'))
);

-- Meetings
CREATE TABLE meetings (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id         UUID REFERENCES organizations(id),
  title          TEXT NOT NULL CHECK (char_length(title) <= 200),
  organizer_id   UUID REFERENCES users(id),
  meeting_range  TSTZRANGE NOT NULL,
  flow_type      TEXT CHECK (flow_type IN ('async','organizer_pick','poll')) NOT NULL,
  status         TEXT CHECK (status IN (
                   'draft','pending','confirmed',
                   'needs_rescheduling','cancelled','rescheduled'
                 )) DEFAULT 'pending',
  notes          TEXT,
  recurrence_id  UUID REFERENCES recurrence_templates(id),
  idempotency_key TEXT UNIQUE,
  created_at     TIMESTAMPTZ DEFAULT now(),
  updated_at     TIMESTAMPTZ DEFAULT now()
);

-- Meeting attendees
CREATE TABLE meeting_attendees (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  meeting_id   UUID REFERENCES meetings(id) ON DELETE CASCADE,
  user_id      UUID REFERENCES users(id),
  is_guest     BOOLEAN DEFAULT FALSE,
  guest_email  TEXT,
  role         TEXT CHECK (role IN ('organizer','required','optional')) NOT NULL,
  response     TEXT CHECK (response IN ('pending','accepted','declined','countered')) DEFAULT 'pending',
  response_note TEXT,
  responded_at TIMESTAMPTZ
);

-- Poll options
CREATE TABLE poll_options (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  meeting_id   UUID REFERENCES meetings(id) ON DELETE CASCADE,
  slot_range   TSTZRANGE NOT NULL,
  created_at   TIMESTAMPTZ DEFAULT now()
);

-- Poll votes
CREATE TABLE poll_votes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  option_id    UUID REFERENCES poll_options(id) ON DELETE CASCADE,
  user_id      UUID REFERENCES users(id),
  vote         TEXT CHECK (vote IN ('yes','if_needed','no')) NOT NULL,
  created_at   TIMESTAMPTZ DEFAULT now(),
  UNIQUE (option_id, user_id)
);

-- Counter proposals
CREATE TABLE counter_proposals (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  meeting_id    UUID REFERENCES meetings(id) ON DELETE CASCADE,
  proposed_by   UUID REFERENCES users(id),
  proposed_range TSTZRANGE NOT NULL,
  note          TEXT,
  status        TEXT CHECK (status IN ('pending','accepted','rejected')) DEFAULT 'pending',
  round         INT DEFAULT 1,
  created_at    TIMESTAMPTZ DEFAULT now()
);

-- Guest links
CREATE TABLE guest_links (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  meeting_id   UUID REFERENCES meetings(id) ON DELETE CASCADE,
  token        TEXT UNIQUE NOT NULL,
  guest_email  TEXT,
  expires_at   TIMESTAMPTZ NOT NULL,
  used_at      TIMESTAMPTZ,
  created_at   TIMESTAMPTZ DEFAULT now()
);

-- Calendar sync credentials
CREATE TABLE calendar_connections (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
  provider        TEXT CHECK (provider IN ('google','outlook')) NOT NULL,
  access_token    TEXT NOT NULL,
  refresh_token   TEXT NOT NULL,
  token_expiry    TIMESTAMPTZ NOT NULL,
  sync_interval   INT DEFAULT 15,
  last_synced_at  TIMESTAMPTZ,
  sync_cursor     TEXT,
  is_active       BOOLEAN DEFAULT TRUE,
  UNIQUE (user_id, provider)
);

-- Sync conflicts
CREATE TABLE sync_conflicts (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID REFERENCES users(id),
  kei_meeting_id    UUID REFERENCES meetings(id),
  external_event_id TEXT NOT NULL,
  external_title    TEXT,
  external_range    TSTZRANGE NOT NULL,
  resolution        TEXT CHECK (resolution IN ('keep_kei','keep_external','keep_both')),
  resolved_at       TIMESTAMPTZ,
  created_at        TIMESTAMPTZ DEFAULT now()
);

-- Recurrence templates
CREATE TABLE recurrence_templates (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  pattern          TEXT CHECK (pattern IN ('daily','weekly','biweekly','monthly')) NOT NULL,
  start_date       DATE NOT NULL,
  end_date         DATE,
  occurrence_count INT,
  meeting_template JSONB NOT NULL,
  created_at       TIMESTAMPTZ DEFAULT now()
);

-- Notifications
CREATE TABLE notifications (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
  event_type   TEXT NOT NULL,
  payload      JSONB NOT NULL,
  read_at      TIMESTAMPTZ,
  created_at   TIMESTAMPTZ DEFAULT now()
);

-- Webhook deliveries
CREATE TABLE webhook_deliveries (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  notification_id UUID REFERENCES notifications(id),
  channel         TEXT CHECK (channel IN ('slack','email','webhook')) NOT NULL,
  url             TEXT,
  status          TEXT CHECK (status IN ('pending','delivered','failed')) DEFAULT 'pending',
  attempts        INT DEFAULT 0,
  last_attempt_at TIMESTAMPTZ,
  response_code   INT,
  created_at      TIMESTAMPTZ DEFAULT now()
);

-- Blocked days
CREATE TABLE blocked_days (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id       UUID REFERENCES organizations(id) ON DELETE CASCADE,
  blocked_date DATE NOT NULL,
  reason       TEXT,
  created_by   UUID REFERENCES users(id),
  created_at   TIMESTAMPTZ DEFAULT now(),
  UNIQUE (org_id, blocked_date)
);

-- Audit log
CREATE TABLE audit_log (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id       UUID REFERENCES organizations(id),
  actor_id     UUID REFERENCES users(id),
  action       TEXT NOT NULL,
  target_type  TEXT,
  target_id    UUID,
  metadata     JSONB,
  created_at   TIMESTAMPTZ DEFAULT now()
);
```

---

# Appendix B — API Surface

## Phase 1

```
POST   /api/auth/setup
POST   /api/auth/login
POST   /api/auth/refresh
DELETE /api/auth/logout
POST   /api/auth/password-reset/request
POST   /api/auth/password-reset/confirm

GET    /api/users
POST   /api/users/invite
GET    /api/users/:id
PUT    /api/users/:id
PUT    /api/users/:id/role
DELETE /api/users/:id

GET    /api/availability/:user_id
PUT    /api/availability/:user_id
GET    /api/availability/:user_id/recurring
PUT    /api/availability/:user_id/recurring
POST   /api/availability/:user_id/import

GET    /api/team/availability
GET    /api/team/overlap

GET    /api/settings
PUT    /api/settings
GET    /api/settings/blocked-days
POST   /api/settings/blocked-days
DELETE /api/settings/blocked-days/:id

WS     /ws/availability
```

## Phase 2 (adds)

```
GET    /api/meetings
POST   /api/meetings
GET    /api/meetings/:id
PUT    /api/meetings/:id/respond
PUT    /api/meetings/:id/cancel
PUT    /api/meetings/:id/reschedule
GET    /api/meetings/:id/history

POST   /api/meetings/:id/poll/vote
GET    /api/meetings/:id/poll/results

GET    /api/suggest?attendees=&duration=&from=&to=

GET    /api/notifications
PUT    /api/notifications/:id/read
PUT    /api/notifications/read-all
GET    /api/notifications/preferences
PUT    /api/notifications/preferences

GET    /api/guest/:token
POST   /api/guest/:token/respond

WS     /ws/meetings
WS     /ws/notifications
```

## Phase 3 (adds)

```
GET    /api/sync/connections
POST   /api/sync/connect/google
POST   /api/sync/connect/outlook
DELETE /api/sync/disconnect/:provider
POST   /api/sync/trigger
GET    /api/sync/status
GET    /api/sync/conflicts
PUT    /api/sync/conflicts/:id/resolve

GET    /api/analytics/me
GET    /api/analytics/team
```

---

# Appendix C — Non-Functional Requirements

| Requirement | Phase 1–2 | Phase 3–4 |
|---|---|---|
| API response time (p95) | < 300ms local | < 200ms |
| Slot suggestion latency | — | < 500ms (50 members) |
| WebSocket latency | < 100ms | < 100ms |
| Concurrent users | 50 | 500 |
| Uptime target | 99% | 99.5% |
| Data retention default | 12 months | Configurable |
| Backup | Daily automated | Daily + WAL archiving |
| Internet dependency | Zero for core features | Zero for core features |
| Min server spec (self-hosted) | 1 vCPU, 512MB RAM | 2 vCPU, 2GB RAM |
| Supported browsers | Chrome, Firefox, Edge (latest 2) | + Safari |
| RTL support | Full from day one | Full |
| Jalali calendar | Full from day one | Full |
| i18n languages | English + Persian | Extensible |

---

*Kei — كي — "When?"*  
*Keep this document in the repo root as `PRODUCT.md`. Update it before changing scope, not after.*
