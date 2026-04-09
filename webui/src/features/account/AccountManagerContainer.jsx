import { useState, useEffect } from 'react'
import { useI18n } from '../../i18n'
import { useAccountsData } from './useAccountsData'
import { useAccountActions } from './useAccountActions'
import QueueCards from './QueueCards'
import ApiKeysPanel from './ApiKeysPanel'
import AccountsTable from './AccountsTable'
import AddKeyModal from './AddKeyModal'
import AddAccountModal from './AddAccountModal'
import QwenAccountsTable from './QwenAccountsTable'
import AddQwenAccountModal from './AddQwenAccountModal'

export default function AccountManagerContainer({ config, onRefresh, onMessage, authFetch }) {
    const { t } = useI18n()
    const apiFetch = authFetch || fetch

    const {
        queueStatus,
        qwenQueueStatus,
        keysExpanded,
        setKeysExpanded,
        accounts,
        page,
        pageSize,
        totalPages,
        totalAccounts,
        loadingAccounts,
        fetchAccounts,
        changePageSize,
        resolveAccountIdentifier,
        searchQuery,
        handleSearchChange,
    } = useAccountsData({ apiFetch })

    // Qwen account state
    const [qwenAccounts, setQwenAccounts] = useState([])
    const [loadingQwen, setLoadingQwen] = useState(false)
    const [showAddQwen, setShowAddQwen] = useState(false)
    const [newQwenAccount, setNewQwenAccount] = useState({ ticket: '', label: '' })
    const [testingQwen, setTestingQwen] = useState({})

    const fetchQwenAccounts = async () => {
        setLoadingQwen(true)
        try {
            const res = await apiFetch('/admin/qwen-accounts')
            if (res.ok) {
                const data = await res.json()
                setQwenAccounts(data.items || [])
            }
        } catch (e) { console.error('fetchQwenAccounts error', e) }
        finally { setLoadingQwen(false) }
    }

    const addQwenAccount = async () => {
        if (!newQwenAccount.ticket.trim()) return
        try {
            const res = await apiFetch('/admin/qwen-accounts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(newQwenAccount),
            })
            if (res.ok) {
                setShowAddQwen(false)
                setNewQwenAccount({ ticket: '', label: '' })
                fetchQwenAccounts()
            }
        } catch (e) { console.error('addQwen error', e) }
    }

    const deleteQwenAccount = async (label) => {
        try {
            const res = await apiFetch(`/admin/qwen-accounts/${encodeURIComponent(label)}`, { method: 'DELETE' })
            if (res.ok) {
                setQwenAccounts(prev => prev.filter(qa => qa.label !== label))
            }
        } catch (e) { console.error('deleteQwen error', e) }
    }

    const testQwenAccount = async (label) => {
        setTestingQwen(prev => ({ ...prev, [label]: true }))
        try {
            const res = await apiFetch(`/admin/qwen-accounts/${encodeURIComponent(label)}/test`, { method: 'POST' })
            if (res.ok) {
                onMessage?.(t('accountManager.qwenTestSuccess'), 'success')
            } else {
                onMessage?.(t('accountManager.testFailed'), 'error')
            }
        } catch (e) { onMessage?.(t('accountManager.testFailed'), 'error') }
        finally { setTestingQwen(prev => ({ ...prev, [label]: false })) }
    }

    // Fetch Qwen accounts on mount
    useEffect(() => { fetchQwenAccounts() }, [])

    const {
        showAddKey,
        setShowAddKey,
        showAddAccount,
        setShowAddAccount,
        newKey,
        setNewKey,
        copiedKey,
        setCopiedKey,
        newAccount,
        setNewAccount,
        loading,
        testing,
        testingAll,
        batchProgress,
        sessionCounts,
        deletingSessions,
        addKey,
        deleteKey,
        addAccount,
        deleteAccount,
        testAccount,
        testAllAccounts,
        deleteAllSessions,
    } = useAccountActions({
        apiFetch,
        t,
        onMessage,
        onRefresh,
        config,
        fetchAccounts,
        resolveAccountIdentifier,
    })

    return (
        <div className="space-y-6">
            {Boolean(config?.env_source_present) && (
                <div className={`rounded-xl border px-4 py-3 text-sm ${
                    config?.env_writeback_enabled
                        ? (config?.env_backed ? 'border-amber-500/30 bg-amber-500/10 text-amber-600' : 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600')
                        : 'border-amber-500/30 bg-amber-500/10 text-amber-600'
                }`}>
                    <p className="font-medium">
                        {config?.env_writeback_enabled
                            ? (config?.env_backed
                                ? t('accountManager.envModeWritebackPendingTitle')
                                : t('accountManager.envModeWritebackActiveTitle'))
                            : t('accountManager.envModeRiskTitle')}
                    </p>
                    <p className="mt-1 text-xs opacity-90">
                        {config?.env_writeback_enabled
                            ? t('accountManager.envModeWritebackDesc', { path: config?.config_path || 'config.json' })
                            : t('accountManager.envModeRiskDesc')}
                    </p>
                </div>
            )}

            <QueueCards queueStatus={queueStatus} qwenQueueStatus={qwenQueueStatus} t={t} />

            <ApiKeysPanel
                t={t}
                config={config}
                keysExpanded={keysExpanded}
                setKeysExpanded={setKeysExpanded}
                setShowAddKey={setShowAddKey}
                copiedKey={copiedKey}
                setCopiedKey={setCopiedKey}
                onDeleteKey={deleteKey}
            />

            <AccountsTable
                t={t}
                accounts={accounts}
                loadingAccounts={loadingAccounts}
                testing={testing}
                testingAll={testingAll}
                batchProgress={batchProgress}
                sessionCounts={sessionCounts}
                deletingSessions={deletingSessions}
                totalAccounts={totalAccounts}
                page={page}
                pageSize={pageSize}
                totalPages={totalPages}
                resolveAccountIdentifier={resolveAccountIdentifier}
                onTestAll={testAllAccounts}
                onShowAddAccount={() => setShowAddAccount(true)}
                onTestAccount={testAccount}
                onDeleteAccount={deleteAccount}
                onDeleteAllSessions={deleteAllSessions}
                onPrevPage={() => fetchAccounts(page - 1)}
                onNextPage={() => fetchAccounts(page + 1)}
                onPageSizeChange={changePageSize}
                searchQuery={searchQuery}
                onSearchChange={handleSearchChange}
                envBacked={Boolean(config?.env_backed)}
            />

            <AddKeyModal
                show={showAddKey}
                t={t}
                newKey={newKey}
                setNewKey={setNewKey}
                loading={loading}
                onClose={() => setShowAddKey(false)}
                onAdd={addKey}
            />

            <AddAccountModal
                show={showAddAccount}
                t={t}
                newAccount={newAccount}
                setNewAccount={setNewAccount}
                loading={loading}
                onClose={() => setShowAddAccount(false)}
                onAdd={addAccount}
            />

            <QwenAccountsTable
                t={t}
                accounts={qwenAccounts}
                loading={loadingQwen}
                testing={testingQwen}
                onShowAddQwen={() => setShowAddQwen(true)}
                onTestQwen={testQwenAccount}
                onDeleteQwen={deleteQwenAccount}
            />

            <AddQwenAccountModal
                show={showAddQwen}
                t={t}
                newQwenAccount={newQwenAccount}
                setNewQwenAccount={setNewQwenAccount}
                onClose={() => setShowAddQwen(false)}
                onAdd={addQwenAccount}
            />
        </div>
    )
}
