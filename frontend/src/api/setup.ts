/**
 * Setup API endpoints
 */
import axios, { type InternalAxiosRequestConfig } from 'axios'

// Create a separate client for setup endpoints (not under /api/v1)
const setupClient = axios.create({
  baseURL: '',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

const SETUP_TOKEN_STORAGE_KEY = 'zsyq_setup_token'

export function initializeSetupTokenFromURL(): boolean {
  const token = new URLSearchParams(window.location.search).get('token')?.trim()
  if (!token) return false
  sessionStorage.setItem(SETUP_TOKEN_STORAGE_KEY, token)
  return true
}

function getSetupToken(): string {
  return sessionStorage.getItem(SETUP_TOKEN_STORAGE_KEY) || ''
}

setupClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getSetupToken()
  if (token && config.headers) {
    config.headers.set('X-Setup-Token', token)
  }
  return config
})

export interface SetupStatus {
  needs_setup: boolean
  step: string
  token_required?: boolean
}

export interface DatabaseConfig {
  host: string
  port: number
  user: string
  password: string
  dbname: string
  sslmode: string
}

export interface RedisConfig {
  host: string
  port: number
  password: string
  db: number
  enable_tls: boolean
}

export interface AdminConfig {
  email: string
  password: string
}

export interface ServerConfig {
  host: string
  port: number
  mode: string
}

export interface InstallRequest {
  database: DatabaseConfig
  redis: RedisConfig
  admin: AdminConfig
  server: ServerConfig
}

export interface InstallResponse {
  message: string
  restart: boolean
}

/**
 * Get setup status
 */
export async function getSetupStatus(): Promise<SetupStatus> {
  const response = await setupClient.get('/setup/status')
  return response.data.data
}

/**
 * Test database connection
 */
export async function testDatabase(config: DatabaseConfig): Promise<void> {
  await setupClient.post('/setup/test-db', config)
}

/**
 * Test Redis connection
 */
export async function testRedis(config: RedisConfig): Promise<void> {
  await setupClient.post('/setup/test-redis', config)
}

/**
 * Perform installation
 */
export async function install(config: InstallRequest): Promise<InstallResponse> {
  const response = await setupClient.post('/setup/install', config)
  return response.data.data
}
