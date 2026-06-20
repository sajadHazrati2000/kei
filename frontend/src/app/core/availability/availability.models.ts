// ─── Availability domain models ───────────────────────────────────────────────

export type SlotStatus = 'free' | 'busy' | 'focus';

/** A single availability block returned by the backend. */
export interface Slot {
  id:           string;
  user_id:      string;
  from:         string; // ISO 8601 UTC
  to:           string; // ISO 8601 UTC
  status:       SlotStatus;
  is_override:  boolean;
  created_at:   string;
}

/** Payload sent when creating or replacing slots. */
export interface SlotInput {
  from:   string; // ISO 8601 UTC
  to:     string; // ISO 8601 UTC
  status: SlotStatus;
}

/** Request body for PUT /api/v1/availability/:user_id */
export interface SetSlotsRequest {
  from:  string;       // window start (ISO 8601)
  to:    string;       // window end   (ISO 8601)
  slots: SlotInput[];
}

/** One member row in the team availability response. */
export interface UserAvailability {
  user: {
    id:       string;
    name:     string;
    timezone: string;
    role:     string;
  };
  slots: Slot[];
}

/** GET /api/v1/team/overlap response envelope. */
export interface OverlapResult {
  date:          string; // YYYY-MM-DD
  overlap_start: string; // HH:MM
  overlap_end:   string; // HH:MM
  members:       UserAvailability[];
}

/** A recurring template row. */
export interface RecurringTemplate {
  id:          string;
  user_id:     string;
  pattern:     'daily' | 'weekly';
  days_of_week: number[];
  start_time:  string; // HH:MM
  end_time:    string; // HH:MM
  status:      SlotStatus;
  valid_from:  string; // YYYY-MM-DD
  valid_until: string | null;
}

// ─── UI helpers ───────────────────────────────────────────────────────────────

/** Per-day dot status derived from the week's slots. */
export type DotStatus = 'none' | 'free' | 'focus' | 'busy';

export interface WeekDay {
  date:   Date;
  dot:    DotStatus;
  slots:  Slot[];
}
