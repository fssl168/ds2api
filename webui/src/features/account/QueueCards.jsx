import { CheckCircle2, Server, ShieldCheck, Globe } from 'lucide-react'

export default function QueueCards({ queueStatus, qwenQueueStatus, t }) {
    return (
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4">
            {/* DeepSeek Stats */}
            <div className="bg-card border border-border rounded-xl p-4 flex flex-col justify-between shadow-sm relative overflow-hidden group">
                <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
                    <CheckCircle2 className="w-16 h-16" />
                </div>
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-widest">{t('accountManager.available')}</p>
                <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-3xl font-bold text-foreground">{queueStatus?.available || 0}</span>
                    <span className="text-xs text-muted-foreground">{t('accountManager.accountsUnit')}</span>
                </div>
            </div>
            <div className="bg-card border border-border rounded-xl p-4 flex flex-col justify-between shadow-sm relative overflow-hidden group">
                <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
                    <Server className="w-16 h-16" />
                </div>
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-widest">{t('accountManager.inUse')}</p>
                <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-3xl font-bold text-foreground">{queueStatus?.in_use || 0}</span>
                    <span className="text-xs text-muted-foreground">{t('accountManager.threadsUnit')}</span>
                </div>
            </div>
            <div className="bg-card border border-border rounded-xl p-4 flex flex-col justify-between shadow-sm relative overflow-hidden group">
                <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
                    <ShieldCheck className="w-16 h-16" />
                </div>
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-widest">{t('accountManager.totalPool')}</p>
                <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-3xl font-bold text-foreground">{queueStatus?.total || 0}</span>
                    <span className="text-xs text-muted-foreground">{t('accountManager.accountsUnit')}</span>
                </div>
            </div>

            {/* Qwen Stats */}
            <div className="bg-card border border-border rounded-xl p-4 flex flex-col justify-between shadow-sm relative overflow-hidden group">
                <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
                    <CheckCircle2 className="w-16 h-16" />
                </div>
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-widest">Qwen 可用</p>
                <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-3xl font-bold text-foreground">{qwenQueueStatus?.available || 0}</span>
                    <span className="text-xs text-muted-foreground">{t('accountManager.accountsUnit')}</span>
                </div>
            </div>
            <div className="bg-card border border-border rounded-xl p-4 flex flex-col justify-between shadow-sm relative overflow-hidden group">
                <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
                    <Server className="w-16 h-16" />
                </div>
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-widest">Qwen 使用中</p>
                <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-3xl font-bold text-foreground">{qwenQueueStatus?.in_use || 0}</span>
                    <span className="text-xs text-muted-foreground">{t('accountManager.threadsUnit')}</span>
                </div>
            </div>
            <div className="bg-card border border-border rounded-xl p-4 flex flex-col justify-between shadow-sm relative overflow-hidden group">
                <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
                    <Globe className="w-16 h-16" />
                </div>
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-widest">Qwen 总数</p>
                <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-3xl font-bold text-foreground">{qwenQueueStatus?.total || 0}</span>
                    <span className="text-xs text-muted-foreground">{t('accountManager.accountsUnit')}</span>
                </div>
            </div>
        </div>
    )
}
