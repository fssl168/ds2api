import { useState } from 'react'
import { Check, Copy, Play, Plus, Trash2 } from 'lucide-react'

export default function QwenAccountsTable({
    t,
    accounts = [],
    loading,
    testing = {},
    onShowAddQwen,
    onTestQwen,
    onDeleteQwen,
}) {
    const [copiedLabel, setCopiedLabel] = useState(null)

    const copyLabel = (label) => {
        navigator.clipboard.writeText(label).then(() => {
            setCopiedLabel(label)
            setTimeout(() => setCopiedLabel(null), 1500)
        })
    }

    return (
        <div className="bg-card border border-border rounded-xl overflow-hidden shadow-sm">
            <div className="p-6 border-b border-border flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        <span className="w-2 h-2 rounded-full bg-blue-500 shadow-[0_0_6px_rgba(59,130,246,0.5)]" />
                        {t('accountManager.qwenAccountsTitle')}
                    </h2>
                    <p className="text-sm text-muted-foreground mt-1">{t('accountManager.qwenAccountsDesc')}</p>
                </div>
                <button
                    onClick={onShowAddQwen}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors font-medium text-sm shadow-sm"
                >
                    <Plus className="w-4 h-4" />
                    {t('accountManager.addQwenAccount')}
                </button>
            </div>

            <div className="divide-y divide-border">
                {loading ? (
                    <div className="p-8 text-center text-muted-foreground">{t('actions.loading')}</div>
                ) : accounts.length > 0 ? (
                    accounts.map((qa, i) => (
                        <div key={i} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-muted/50 transition-colors">
                            <div className="flex items-center gap-3 min-w-0">
                                <div className="w-2 h-2 rounded-full bg-blue-500 shadow-[0_0_6px_rgba(59,130,246,0.5)] shrink-0" />
                                <div className="min-w-0">
                                    <div
                                        className="font-medium truncate flex items-center gap-1.5 cursor-pointer hover:text-blue-400 transition-colors group"
                                        onClick={() => copyLabel(qa.label)}
                                    >
                                        <span className="truncate">{qa.label || '-'}</span>
                                        {copiedLabel === qa.label
                                            ? <Check className="w-3 h-3 text-emerald-500 shrink-0" />
                                            : <Copy className="w-3 h-3 opacity-0 group-hover:opacity-50 shrink-0 transition-opacity" />
                                        }
                                    </div>
                                    <div className="font-mono text-xs text-muted-foreground mt-0.5 bg-muted px-2 py-0.5 rounded inline-block truncate max-w-[300px]">
                                        {qa.ticket_preview || qa.ticket}
                                    </div>
                                </div>
                            </div>
                            <div className="flex items-center gap-2 self-start sm:self-auto">
                                <button
                                    onClick={() => onTestQwen(qa.label)}
                                    disabled={testing[qa.label]}
                                    className="px-3 py-1.5 text-xs font-medium border border-border rounded-md hover:bg-secondary transition-colors disabled:opacity-50"
                                >
                                    {testing[qa.label] ? t('actions.testing') : t('actions.test')}
                                </button>
                                <button
                                    onClick={() => onDeleteQwen(qa.label)}
                                    className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-md transition-colors"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    ))
                ) : (
                    <div className="p-8 text-center text-muted-foreground">{t('accountManager.noQwenAccounts')}</div>
                )}
            </div>

            {accounts.length > 0 && (
                <div className="p-4 border-t border-border flex items-center justify-between text-sm text-muted-foreground">
                    <span>{t('accountManager.qwenTotal', { count: accounts.length })}</span>
                </div>
            )}
        </div>
    )
}