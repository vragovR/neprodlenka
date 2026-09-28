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

export type ChildPatch = Partial<ChildInput & Pick<Child, 'active' | 'parent_ids' | 'clubs'>>

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
  marks: Record<string, Status>
  total: number
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

export interface Day {
  date: string
  school: boolean
  published: boolean
  closed: boolean
  children: DayRow[]
  message: BotMessage | null
}

export interface DayClubChange extends DayClub {
  child_id: number
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

export interface ChildReport {
  month: string
  days: ReportDay[]
  child: ReportRow
  price: number
  prev_uid: string
  next_uid: string
}
