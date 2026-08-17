async function request(path, options = {}) {
  const res = await fetch(path, options)
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || `请求失败 (${res.status})`)
  }
  return data
}

export function getHealth() {
  return request('/api/health')
}

export function getProfile() {
  return request('/api/profile')
}

export function updateProfile(profile, token) {
  return request('/api/profile', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(profile),
  })
}

export function getLogs(type) {
  const q = type ? `?type=${encodeURIComponent(type)}` : ''
  return request(`/api/logs${q}`)
}

export function createLog(log, token) {
  return request('/api/logs', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(log),
  })
}
