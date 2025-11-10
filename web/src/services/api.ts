import axios, { AxiosInstance, AxiosError } from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 30000,
    })

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      response => response,
      (error: AxiosError) => {
        console.error('API Error:', error.message)
        return Promise.reject(error)
      },
    )
  }

  get<T>(url: string, config = {}) {
    return this.client.get<T>(url, config)
  }

  post<T>(url: string, data?: any, config = {}) {
    return this.client.post<T>(url, data, config)
  }

  put<T>(url: string, data?: any, config = {}) {
    return this.client.put<T>(url, data, config)
  }

  delete<T>(url: string, config = {}) {
    return this.client.delete<T>(url, config)
  }
}

export const apiClient = new ApiClient()
