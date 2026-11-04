import { request } from './client'
import type {
  Attendance,
  Child,
  ChildInput,
  ChildPatch,
  ChildReport,
  ChildrenResponse,
  Day,
  DayClub,
  MarkChange,
  Month,
  PaymentChange,
  Report,
} from './types'

export const login = (login: string, password: string) =>
  request<{ login: string }>('POST', '/login', { login, password })

export const logout = () => request<void>('DELETE', '/logout')

export const getChildren = () => request<ChildrenResponse>('GET', '/children')

export const addChild = (child: ChildInput) => request<Child>('POST', '/children', child)

export const updateChild = (id: number, patch: ChildPatch) =>
  request<Child>('PATCH', `/children/${id}`, patch)

export const getMonth = (month: string) => request<Month>('GET', `/months/${month}`)

export const putMonth = (month: string, noSchoolDays: string[]) =>
  request<void>('PUT', `/months/${month}`, { no_school_days: noSchoolDays })

export const getAttendance = (month: string) => request<Attendance>('GET', `/attendance/${month}`)

export const putAttendance = (month: string, marks: MarkChange[], payments?: PaymentChange[]) =>
  request<void>('PUT', `/attendance/${month}`, { marks, payments })

export const getDay = (date: string) => request<Day>('GET', `/days/${date}`)

export const putDayClubs = (date: string, childId: number, clubs: DayClub[]) =>
  request<void>('PUT', `/days/${date}/children/${childId}/clubs`, { clubs })

export const getReport = (uid: string) =>
  request<Report>('GET', `/reports/${encodeURIComponent(uid)}`)

export const getChildReport = (uid: string, id: string) =>
  request<ChildReport>(
    'GET',
    `/reports/${encodeURIComponent(uid)}/children/${encodeURIComponent(id)}`,
  )

export const debugSendDay = (date: string) => request<void>('POST', `/debug/day/${date}`, {})

export const debugSendMonth = (month: string) => request<void>('POST', `/debug/month/${month}`, {})

export const debugCloseDay = (date: string) => request<void>('POST', `/debug/close/${date}`, {})

export const debugSummaries = () => request<{ months: string[] }>('GET', '/debug/summaries')

export const debugDeleteDay = (date: string) => request<void>('DELETE', `/debug/day/${date}`)

export const debugDeleteMonth = (month: string) => request<void>('DELETE', `/debug/month/${month}`)
