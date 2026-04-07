import clsx from 'clsx'

import { useI18n } from '../../i18n'
import { useApiTesterState } from './useApiTesterState'
import { useChatStreamClient } from './useChatStreamClient'
import ConfigPanel from './ConfigPanel'
import ChatPanel from './ChatPanel'

export default function ApiTesterContainer({ config, onMessage, authFetch }) {
    const { t } = useI18n()

    const {
        model,
        setModel,
        message,
        setMessage,
        apiKey,
        setApiKey,
        selectedAccount,
        setSelectedAccount,
        response,
        setResponse,
        loading,
        setLoading,
        streamingContent,
        setStreamingContent,
        streamingThinking,
        setStreamingThinking,
        isStreaming,
        setIsStreaming,
        streamingMode,
        setStreamingMode,
        configExpanded,
        setConfigExpanded,
        abortControllerRef,
    } = useApiTesterState({ t })

    const accounts = config.accounts || []
    const resolveAccountIdentifier = (acc) => {
        if (!acc || typeof acc !== 'object') return ''
        return String(acc.identifier || acc.email || acc.mobile || '').trim()
    }
    const configuredKeys = config.keys || []
    const trimmedApiKey = apiKey.trim()
    const defaultKey = configuredKeys[0] || ''
    const effectiveKey = trimmedApiKey || defaultKey
    const customKeyActive = trimmedApiKey !== ''
    const customKeyManaged = customKeyActive && configuredKeys.includes(trimmedApiKey)

    const models = [
        { id: 'qwen-plus', name: 'qwen-plus', icon: 'MessageSquare', desc: t('apiTester.models.qwenPlus'), color: 'text-blue-500' },
        { id: 'qwen-max', name: 'qwen-max', icon: 'Cpu', desc: t('apiTester.models.qwenMax'), color: 'text-blue-600' },
        { id: 'qwen-coder', name: 'qwen-coder', icon: 'Code', desc: t('apiTester.models.qwenCoder'), color: 'text-violet-500' },
        { id: 'qwen-flash', name: 'qwen-flash', icon: 'Zap', desc: t('apiTester.models.qwenFlash'), color: 'text-green-500' },
        { id: 'qwen3.5-plus', name: 'qwen3.5-plus', icon: 'Sparkles', desc: t('apiTester.models.qwen35Plus'), color: 'text-indigo-500' },
        { id: 'qwen3.5-flash', name: 'qwen3.5-flash', icon: 'Bolt', desc: t('apiTester.models.qwen35Flash'), color: 'text-emerald-500' },
        { id: 'deepseek-chat', name: 'deepseek-chat', icon: 'MessageSquare', desc: t('apiTester.models.chat'), color: 'text-amber-500' },
        { id: 'deepseek-reasoner', name: 'deepseek-reasoner', icon: 'Cpu', desc: t('apiTester.models.reasoner'), color: 'text-amber-600' },
        { id: 'deepseek-chat-search', name: 'deepseek-chat-search', icon: 'SearchIcon', desc: t('apiTester.models.chatSearch'), color: 'text-cyan-500' },
        { id: 'deepseek-reasoner-search', name: 'deepseek-reasoner-search', icon: 'SearchIcon', desc: t('apiTester.models.reasonerSearch'), color: 'text-cyan-600' },
    ]

    const { runTest, stopGeneration } = useChatStreamClient({
        t,
        onMessage,
        model,
        message,
        effectiveKey,
        selectedAccount,
        streamingMode,
        abortControllerRef,
        setLoading,
        setIsStreaming,
        setResponse,
        setStreamingContent,
        setStreamingThinking,
    })

    return (
        <div className={clsx('flex flex-col lg:grid lg:grid-cols-12 gap-6 h-[calc(100vh-140px)]')}>
            <ConfigPanel
                t={t}
                configExpanded={configExpanded}
                setConfigExpanded={setConfigExpanded}
                models={models}
                model={model}
                setModel={setModel}
                streamingMode={streamingMode}
                setStreamingMode={setStreamingMode}
                selectedAccount={selectedAccount}
                setSelectedAccount={setSelectedAccount}
                accounts={accounts}
                resolveAccountIdentifier={resolveAccountIdentifier}
                apiKey={apiKey}
                setApiKey={setApiKey}
                config={config}
                customKeyActive={customKeyActive}
                customKeyManaged={customKeyManaged}
            />

            <ChatPanel
                t={t}
                message={message}
                setMessage={setMessage}
                response={response}
                isStreaming={isStreaming}
                loading={loading}
                streamingThinking={streamingThinking}
                streamingContent={streamingContent}
                onRunTest={runTest}
                onStopGeneration={stopGeneration}
                model={model}
            />
        </div>
    )
}
