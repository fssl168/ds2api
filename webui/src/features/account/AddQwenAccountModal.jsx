import { X } from 'lucide-react'

export default function AddQwenAccountModal({
    show,
    t,
    newQwenAccount,
    setNewQwenAccount,
    loading,
    onClose,
    onAdd,
}) {
    if (!show) return null

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-in fade-in">
            <div className="bg-card w-full max-w-md rounded-xl border border-border shadow-2xl overflow-hidden animate-in zoom-in-95">
                <div className="p-4 border-b border-border flex justify-between items-center">
                    <h3 className="font-semibold">{t('accountManager.modalAddQwenTitle')}</h3>
                    <button onClick={onClose} className="text-muted-foreground hover:text-foreground">
                        <X className="w-5 h-5" />
                    </button>
                </div>
                <div className="p-6 space-y-4">
                    <div>
                        <label className="block text-sm font-medium mb-1.5">{t('accountManager.qwenTicketLabel')} <span className="text-destructive">*</span></label>
                        <input
                            type="text"
                            className="input-field font-mono"
                            placeholder="1ogROHqapzX0CcdoyjQj$klhtV2M3wGtaHs5PPt9Kx_F0ty14Q2igUFFPW4ybaNzpf8Gz1g_Byns0"
                            value={newQwenAccount.ticket}
                            onChange={e => setNewQwenAccount({ ...newQwenAccount, ticket: e.target.value })}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1.5">{t('accountManager.qwenLabelOptional')}</label>
                        <input
                            type="text"
                            className="input-field"
                            placeholder={t('accountManager.qwenLabelPlaceholder')}
                            value={newQwenAccount.label}
                            onChange={e => setNewQwenAccount({ ...newQwenAccount, label: e.target.value })}
                        />
                        <p className="text-xs text-muted-foreground mt-1">{t('accountManager.qwenLabelHint')}</p>
                    </div>
                    <div className="flex justify-end gap-2 pt-2">
                        <button onClick={onClose} className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors text-sm font-medium">{t('actions.cancel')}</button>
                        <button onClick={onAdd} disabled={loading} className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors text-sm font-medium disabled:opacity-50">
                            {loading ? t('accountManager.addAccountLoading') : t('accountManager.addQwenAction')}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    )
}