export type Status = 'present' | 'absent'

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

export type ChildPatch = Partial<ChildInput & Pick<Child, 'active' | 'parent_ids'>>

export interface Month {
  month: string
  no_school_days: string[]
  report_url: string
}

export interface AttendanceDay {
  date: string
  published: boolean
  closed: boolean
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
}

export interface MarkChange {
  date: string
  child_id: number
  status: Status | null
}

export interface ReportRow extends PersonName {
  marks: (Status | null)[]
  present: number
  absent: number
  total: number
}

export interface Report {
  month: string
  days: string[]
  children: ReportRow[]
  total: number
}
