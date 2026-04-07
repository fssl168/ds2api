import { useEffect, useState, useCallback } from 'react'
import { Shield, RefreshCw, Clock, User, Terminal } from 'lucide-react'
import clsx from 'clsx'

const ACTION_ICONS = {
    add_key: '🔑',
    delete_key: '🗑️',
    add_account: '👤',
    delete_account: '❌',
    add_qwen_account: '🎫',
    delete_qwen_account: '🚫',
    update_password: '🔒',
    update_config: '⚙️',
    batch_import: '📦',
}

const ACTION_COLORS = {
    add_key: 'text-emerald-400',
    delete_key: 'text-red-400',
    add_account: 'text-blue-400',
    delete_account: 'text-red-400',
    add_qwen_account: 'text-purple-400',
    delete_qwen_account: 'text-red-400',
    update_password: 'text-amber-400',
    update_config: 'text-cyan-400',
    batch_import: 'text-green-400',
}

export default function AuditLogContainer({ authFetch, onMessage }) {
    const [entries, setEntries] = useState([])
    const [loading, setLoading] = useState(false)
    const [autoRefresh, setAutoRefresh] = useState(true)

    const fetchLogs = useCallback(async () => {
        setLoading(true)
        try {
            const res = await authFetch('/admin/audit-log?limit=100')
            if (res.ok) {
                const data = await res.json()
                setEntries(data.entries || [])
            }
        } catch (err) {
            onMessage?.({ type: 'error', text: `Failed to load audit logs: ${err.message}` })
        } finally {
            setLoading(false)
        }
    }, [authFetch, onMessage])

    useEffect(() => {
        fetchLogs()
        if (!autoRefresh) return
        const interval = setInterval(fetchLogs, 10000)
        return () => clearInterval(interval)
    }, [fetchLogs, autoRefresh])

    const formatTime = (iso) => {
        try {
            const d = new Date(iso)
            return d.toLocaleString()
        } catch {
            return iso
        }
    }

    const actionLabel = (action) => {
        return action.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center">
                        <Shield className="w-5 h-5 text-primary" />
                    </div>
                    <div>
                        <h2 className="text-lg font-semibold">Audit Log</h2>
                        <p className="text-xs text-muted-foreground">Recent administrative operations ({entries.length} entries)</p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <button
                        onClick={() => setAutoRefresh(!autoRefresh)}
                        className={clsx(
                            "px-3 py-1.5 rounded-lg text-xs font-medium border transition-all",
                            autoRefresh ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20" : "bg-secondary/50 text-muted-foreground border-border"
                        )}
                    >
                        {autoRefresh ? 'Auto ON' : 'Auto OFF'}
                    </button>
                    <button
                        onClick={fetchLogs}
                        disabled={loading}
                        className="px-3 py-1.5 rounded-lg text-xs font-medium bg-secondary/50 hover:bg-secondary border border-border transition-all flex items-center gap-1.5 disabled:opacity-50"
                    >
                        <RefreshCw className={clsx("w-3.5 h-3.5", loading && "animate-spin")} />
                        Refresh
                    </button>
                </div>
            </div>

            {entries.length === 0 && !loading ? (
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground border border-dashed border-border rounded-xl">
                    <Terminal className="w-12 h-12 mb-3 opacity-20" />
                    <p className="text-sm font-medium">No audit entries yet</p>
                    <p className="text-xs mt-1 opacity-60">Administrative operations will appear here</p>
                </div>
            ) : (
                <div className="border border-border rounded-xl overflow-hidden">
                    <div className="bg-card/50 px-4 py-2.5 border-b border-border grid grid-cols-[1fr_2fr_140px_160px] gap-3 text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                        <span>Action</span>
                        <span>Detail</span>
                        <span className="flex items-center gap-1"><Clock className="w-3 h-3" /> Time</span>
                        <span className="flex items-center gap-1"><User className="w-3 h-3" /> Source</span>
                    </div>
                    <div className="divide-y divide-border max-h-[600px] overflow-y-auto">
                        {entries.map((entry, i) => (
                            <div key={i} className="px-4 py-3 grid grid-cols-[1fr_2fr_140px_160px] gap-3 items-start hover:bg-secondary/30 transition-colors text-sm">
                                <div className="flex items-center gap-2 min-w-0">
                                    <span className="text-base">{ACTION_ICONS[entry.action] || '📋'}</span>
                                    <span className={clsx("font-medium font-mono text-xs truncate", ACTION_COLORS[entry.action] || "text-foreground")}>
                                        {actionLabel(entry.action)}
                                    </span>
                                </div>
                                <span className="text-muted-foreground text-xs truncate" title={entry.detail}>
                                    {entry.detail}
                                </span>
                                <span className="text-xs text-muted-foreground tabular-nums whitespace-nowrap">
                                    {formatTime(entry.time)}
                                </span>
                                <span className="text-xs text-muted-foreground font-mono truncate" title={entry.remote}>
                                    {entry.remote?.replace(/^.*:/, '') || '-'}
                                </span>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}
