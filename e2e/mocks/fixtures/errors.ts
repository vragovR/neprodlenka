export interface ApiError {
  error: string
  message: string
}

export const invalidCredentials: ApiError = { error: 'invalid_credentials', message: 'Неверный логин или пароль' }

export const notFound: ApiError = { error: 'not_found', message: 'Не найдено' }
