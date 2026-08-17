import { useEffect, useState } from 'react'
import { getProfile, updateProfile, getLogs, createLog, getHealth } from './api.js'

const TYPE_LABELS = { work: '工作', study: '学习', daily: '日报', summary: '总结' }
const FILTERS = [
  { value: '', label: '全部' },
  { value: 'work', label: '工作' },
  { value: 'study', label: '学习' },
  { value: 'daily', label: '日报' },
  { value: 'summary', label: '总结' },
]

function formatDate(iso) {
  return new Date(iso).toLocaleString('zh-CN')
}

function ProfileCard({ profile, onSaved }) {
  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState(profile)
  const [token, setToken] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => setForm(profile), [profile])

  function set(field) {
    return (e) => setForm({ ...form, [field]: e.target.value })
  }

  async function save(e) {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      await updateProfile(form, token)
      onSaved()
      setEditing(false)
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  if (!editing) {
    return (
      <section className="card profile">
        <h2>{profile.name}</h2>
        <div className="profile-grid">
          <Info label="电话" value={profile.phone} />
          <Info label="邮箱" value={profile.email} />
          <Info label="技术方向" value={profile.tech_direction} />
          <Info label="学习目标" value={profile.learning_goals} />
        </div>
        <button className="link-btn" onClick={() => setEditing(true)}>
          编辑个人信息
        </button>
      </section>
    )
  }

  return (
    <section className="card profile">
      <h2>编辑个人信息</h2>
      <form onSubmit={save} className="form">
        <label>姓名 <input value={form.name} onChange={set('name')} required /></label>
        <label>电话 <input value={form.phone} onChange={set('phone')} /></label>
        <label>邮箱 <input value={form.email} onChange={set('email')} /></label>
        <label>技术方向 <input value={form.tech_direction} onChange={set('tech_direction')} /></label>
        <label>学习目标 <textarea value={form.learning_goals} onChange={set('learning_goals')} rows={2} /></label>
        <label>Owner token <input type="password" value={token} onChange={(e) => setToken(e.target.value)} required /></label>
        {error && <p className="error">{error}</p>}
        <div className="row">
          <button type="submit" disabled={saving}>{saving ? '保存中…' : '保存'}</button>
          <button type="button" className="ghost" onClick={() => setEditing(false)}>取消</button>
        </div>
      </form>
    </section>
  )
}

function Info({ label, value }) {
  if (!value) return null
  return (
    <div className="info-item">
      <span className="info-label">{label}</span>
      <span>{value}</span>
    </div>
  )
}

function LogForm({ onCreated }) {
  const [open, setOpen] = useState(false)
  const [type, setType] = useState('work')
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [token, setToken] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      await createLog({ type, title, content }, token)
      setTitle('')
      setContent('')
      setOpen(false)
      onCreated()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  if (!open) {
    return (
      <button className="primary" onClick={() => setOpen(true)}>
        写日志
      </button>
    )
  }

  return (
    <section className="card">
      <h2>写日志</h2>
      <form onSubmit={submit} className="form">
        <label>
          类型
          <select value={type} onChange={(e) => setType(e.target.value)}>
            {Object.entries(TYPE_LABELS).map(([k, v]) => (
              <option key={k} value={k}>{v}</option>
            ))}
          </select>
        </label>
        <label>标题 <input value={title} onChange={(e) => setTitle(e.target.value)} required /></label>
        <label>正文 <textarea value={content} onChange={(e) => setContent(e.target.value)} rows={5} required /></label>
        <label>Owner token <input type="password" value={token} onChange={(e) => setToken(e.target.value)} required /></label>
        {error && <p className="error">{error}</p>}
        <div className="row">
          <button type="submit" disabled={saving}>{saving ? '提交中…' : '提交'}</button>
          <button type="button" className="ghost" onClick={() => setOpen(false)}>取消</button>
        </div>
      </form>
    </section>
  )
}

function LogList({ logs }) {
  const [expanded, setExpanded] = useState(null)

  if (logs.length === 0) {
    return <p className="empty">暂无日志</p>
  }

  return (
    <div className="log-list">
      {logs.map((log) => (
        <article key={log.id} className="card log" onClick={() => setExpanded(expanded === log.id ? null : log.id)}>
          <div className="log-head">
            <span className={`badge badge-${log.type}`}>{TYPE_LABELS[log.type] || log.type}</span>
            <h3>{log.title}</h3>
            <time>{formatDate(log.created_at)}</time>
          </div>
          {expanded === log.id && <p className="log-content">{log.content}</p>}
        </article>
      ))}
    </div>
  )
}

export default function App() {
  const [profile, setProfile] = useState(null)
  const [logs, setLogs] = useState([])
  const [filter, setFilter] = useState('')
  const [error, setError] = useState('')
  const [online, setOnline] = useState(false)

  async function load() {
    try {
      const [h, p, l] = await Promise.all([getHealth(), getProfile(), getLogs(filter)])
      setOnline(h.status === 'ok')
      setProfile(p)
      setLogs(l)
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    load()
  }, [filter])

  return (
    <div className="page">
      <header className="header">
        <h1>个人主页与成长日志</h1>
        <span className={online ? 'status on' : 'status off'}>
          {online ? '● 服务在线' : '○ 服务离线'}
        </span>
      </header>

      {error && <p className="error banner">{error}</p>}

      {profile && <ProfileCard profile={profile} onSaved={load} />}

      <div className="toolbar">
        <LogForm onCreated={load} />
      </div>

      <div className="filters">
        {FILTERS.map((f) => (
          <button
            key={f.value}
            className={filter === f.value ? 'chip active' : 'chip'}
            onClick={() => setFilter(f.value)}
          >
            {f.label}
          </button>
        ))}
      </div>

      <LogList logs={logs} />
    </div>
  )
}
