export type Status = 'pending' | 'present' | 'absent'

export interface PersonName {
  last_name: string
  first_name: string
  middle_name: string
}

export interface Child extends PersonName {
  id: number
  price: number
  active: boolean
  weekdays: number[]
  parent_ids: number[]
  clubs: Club[]
}

export interface Club {
  weekday: number
  start: string
  name: string
}

export interface Member {
  user_id: number
  name: string
}

export interface ChildrenResponse {
  children: Child[]
  members: Member[]
}

export interface ChildInput extends PersonName {
  price: number
}

export type ChildPatch = Partial<
  ChildInput & Pick<Child, 'active' | 'weekdays' | 'parent_ids' | 'clubs'>
>

export interface Month {
  month: string
  no_school_days: string[]
  report_url: string
}

export interface AttendanceDay {
  date: string
  published: boolean
  closed: boolean
  school: boolean
}

export interface AttendanceRow extends PersonName {
  child_id: number
  active: boolean
  weekdays: number[]
  marks: Record<string, Status>
  total: number
  paid: boolean
}

export interface Attendance {
  month: string
  days: AttendanceDay[]
  children: AttendanceRow[]
  message: BotMessage | null
  sent: boolean
}

export interface DayClub {
  start: string
  name: string
}

export interface DayRow extends PersonName {
  child_id: number
  active: boolean
  status: Status | null
  clubs: DayClub[]
}

export interface BotMessage {
  text: string
  buttons: string[]
}

export interface ClubChange {
  name: string
  start: string | null
}

export interface DayClubChange extends ClubChange {
  child_id: number
  last_name: string
  first_name: string
}

export interface Day {
  date: string
  school: boolean
  published: boolean
  closed: boolean
  children: DayRow[]
  message: BotMessage | null
  changes: DayClubChange[]
  changes_editable: boolean
}

export interface PaymentChange {
  child_id: number
  paid: boolean
}

export interface MarkChange {
  date: string
  child_id: number
  status: Status | null
}

export interface ReportRow extends PersonName {
  id: number
  marks: (Status | null)[]
  present: number
  absent: number
  total: number
}

export interface ReportDay {
  date: string
  school: boolean
}

export interface Report {
  month: string
  days: ReportDay[]
  children: ReportRow[]
  total: number
}

export interface ChildClubChange extends ClubChange {
  date: string
}

export interface ChildReport {
  month: string
  days: ReportDay[]
  child: ReportRow
  price: number
  clubs: Club[]
  club_changes: ChildClubChange[]
}
