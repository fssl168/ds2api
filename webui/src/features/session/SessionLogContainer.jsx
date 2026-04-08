import { useEffect, useState, useCallback } from 'react'
import { History, RefreshCw, Clock, User, Zap, Filter, Globe, AlertCircle, CheckCircle } from 'lucide-react'
import clsx from 'clsx'

const ENGINE_ICONS = {
	deepseek: '🔵',
	qwen: '🟢',
	gemini: '🟡',
	claude: '🟠',
}

const ENGINE_LABELS = {
	deepseek: 'DeepSeek',
	qwen: 'Qwen',
	gemini: 'Gemini',
	claude: 'Claude',
}

const STATUS_CONFIG = {
	success: { icon: CheckCircle, color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', label: 'Success' },
	error: { icon: AlertCircle, color: 'text-red-400', bg: 'bg-red-500/10', border: 'border-red-500/20', label: 'Error' },
}

export default function SessionLogContainer({ authFetch, onMessage }) {
	const [entries, setEntries] = useState([])
	const [loading, setLoading] = useState(false)
	const [autoRefresh, setAutoRefresh] = useState(false)
	const [engineFilter, setEngineFilter] = useState('')
	const [statusFilter, setStatusFilter] = useState('')

	const fetchLogs = useCallback(async () => {
		setLoading(true)
		try {
			const params = new URLSearchParams({ limit: '200' })
			if (engineFilter) params.set('engine', engineFilter)
			if (statusFilter) params.set('status', statusFilter)
			const res = await authFetch(`/admin/session-logs?${params.toString()}`)
			if (res.ok) {
				const data = await res.json()
				setEntries(data.entries || [])
			}
		} catch (err) {
			onMessage?.({ type: 'error', text: `Failed to load session logs: ${err.message}` })
		} finally {
			setLoading(false)
		}
	}, [authFetch, onMessage, engineFilter, statusFilter])

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

	const formatLatency = (ms) => {
		if (!ms && ms !== 0) return '-'
		if (ms < 1000) return `${ms}ms`
		return `${(ms / 1000).toFixed(1)}s`
	}

	const maskCaller = (id) => {
		if (!id || id.length <= 12) return id || '-'
		return id.slice(0, 10) + '...'
	}

	const maskRemote = (addr) => {
		if (!addr) return '-'
		const parts = addr.lastIndexOf(':')
		if (parts <= 0) return addr
		return addr.slice(0, Math.max(parts - 2, 0)) + '**'
	}

	const statusCounts = entries.reduce((acc, e) => {
		acc[e.status] = (acc[e.status] || 0) + 1
		acc['_total'] = (acc['_total'] || 0) + 1
		return acc
	}, {})

	const engineCounts = entries.reduce((acc, e) => {
		const key = e.engine || 'unknown'
		acc[key] = (acc[key] || 0) + 1
		return acc
	}, {})

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-3">
					<div className="w-10 h-10 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center">
						<History className="w-5 h-5 text-primary" />
					</div>
					<div>
						<h2 className="text-lg font-semibold">Session Logs</h2>
						<p className="text-xs text-muted-foreground">
							All engine requests ({entries.length} entries)
							{statusCounts._total > 0 && (
								<span> · <span className="text-emerald-400">{statusCounts.success || 0}</span> ok / <span className="text-red-400">{statusCounts.error || 0}</span> err</span>
							)}
						</p>
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

			<div className="flex flex-wrap items-center gap-2 p-3 bg-card/50 rounded-lg border border-border">
				<Filter className="w-3.5 h-3.5 text-muted-foreground" />
				<span className="text-xs text-muted-foreground font-medium">Engine:</span>
				{['', 'deepseek', 'qwen', 'gemini', 'claude'].map((eng) => (
					<button
						key={eng || 'all'}
						onClick={() => setEngineFilter(eng)}
						className={clsx(
							"px-2.5 py-1 rounded-md text-xs font-medium border transition-all",
							engineFilter === eng
								? "bg-primary/10 text-primary border-primary/30"
								: "bg-secondary/50 text-muted-foreground border-border hover:bg-secondary"
						)}
					>
						{eng ? `${ENGINE_ICONS[eng] || ''} ${ENGINE_LABELS[eng] || eng}` : 'All'}
						{engineCounts[eng] !== undefined ? ` (${engineCounts[eng]})` : ''}
					</button>
				))}
				<span className="text-xs text-muted-foreground mx-1">|</span>
				<span className="text-xs text-muted-foreground font-medium">Status:</span>
				{['', 'success', 'error'].map((st) => (
					<button
						key={st || 'all'}
						onClick={() => setStatusFilter(st)}
						className={clsx(
							"px-2.5 py-1 rounded-md text-xs font-medium border transition-all",
							statusFilter === st
								? "bg-primary/10 text-primary border-primary/30"
								: "bg-secondary/50 text-muted-foreground border-border hover:bg-secondary"
						)}
					>
						{st ? STATUS_CONFIG[st]?.label || st : 'All'}
					</button>
				))}
			</div>

			{entries.length === 0 && !loading ? (
				<div className="flex flex-col items-center justify-center py-16 text-muted-foreground border border-dashed border-border rounded-xl">
					<History className="w-12 h-12 mb-3 opacity-20" />
					<p className="text-sm font-medium">No session records yet</p>
					<p className="text-xs mt-1 opacity-60">API requests will appear here</p>
				</div>
			) : (
				<div className="border border-border rounded-xl overflow-hidden">
					<div className="bg-card/50 px-4 py-2.5 border-b border-border grid grid-cols-[140px_1fr_80px_70px_90px_80px_120px] gap-3 text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
						<span>Time</span>
						<span>Model</span>
						<span>Engine</span>
						<span>Status</span>
						<span>Latency</span>
						<span>Messages</span>
						<span className="flex items-center gap-1"><User className="w-3 h-3" /> Caller</span>
					</div>
					<div className="divide-y divide-border max-h-[600px] overflow-y-auto">
						{entries.map((entry, i) => {
							const sc = STATUS_CONFIG[entry.status] || STATUS_CONFIG.success
							const StatusIcon = sc.icon
							return (
								<div key={i} className="px-4 py-2.5 grid grid-cols-[140px_1fr_80px_70px_90px_80px_120px] gap-3 items-start hover:bg-secondary/30 transition-colors text-sm">
									<span className="text-xs text-muted-foreground tabular-nums whitespace-nowrap">
										{formatTime(entry.time)}
									</span>
									<span className="text-xs font-mono font-medium truncate" title={entry.model}>
										{entry.model || '-'}
									</span>
									<span className={clsx("text-xs font-medium flex items-center gap-1")}>
										<span>{ENGINE_ICONS[entry.engine] || '❓'}</span>
										<span>{ENGINE_LABELS[entry.engine] || entry.engine}</span>
									</span>
									<span className={clsx("inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold", sc.bg, sc.border, sc.color)}>
										<StatusIcon className="w-3 h-3" />
										{sc.label}
									</span>
									<span className={clsx("text-xs tabular-nums font-mono font-medium", entry.latency_ms > 5000 ? "text-amber-400" : "text-foreground")}>
										{formatLatency(entry.latency_ms)}
									</span>
									<span className="text-xs text-muted-foreground tabular-nums text-center">
										{entry.message_count ?? '-'}
										{entry.is_stream && <span className="ml-0.5 text-[9px] text-blue-400">(S)</span>}
									</span>
									<span className="text-xs text-muted-foreground font-mono truncate" title={entry.caller_id}>
										{maskCaller(entry.caller_id)}
									</span>
								</div>
							)
						})}
					</div>
				</div>
			)}

			{entries.some(e => e.status === 'error' && e.error_msg) && (
				<div className="mt-2 p-3 bg-red-500/5 border border-red-500/10 rounded-lg">
					<div className="text-[11px] font-bold uppercase tracking-wider text-red-400 mb-2 flex items-center gap-1.5">
						<AlertCircle className="w-3.5 h-3.5" /> Recent Errors
					</div>
					<div className="space-y-1 max-h-40 overflow-y-auto">
						{entries.filter(e => e.status === 'error' && e.error_msg).slice(0, 5).map((e, i) => (
							<div key={i} className="text-xs text-red-300/80 font-mono bg-red-500/5 px-2 py-1 rounded truncate" title={e.error_msg}>
								[{formatTime(e.time)}] {e.model}: {e.error_msg}
							</div>
						))}
					</div>
				</div>
			)}
		</div>
	)
}
