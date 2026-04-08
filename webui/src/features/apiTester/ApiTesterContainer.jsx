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
    const qwenAccounts = config.qwen_accounts || []
    const resolveAccountIdentifier = (acc) => {
        if (!acc || typeof acc !== 'object') return ''
        return String(acc.identifier || acc.email || acc.mobile || '').trim()
    }
    
    // Select accounts based on model type
    const isQwenModel = model.startsWith('qwen')
    const displayAccounts = isQwenModel ? qwenAccounts : accounts
    const configuredKeys = config.keys || []
    const trimmedApiKey = apiKey.trim()
    const defaultKey = configuredKeys[0] || ''
    const effectiveKey = trimmedApiKey || defaultKey
    const customKeyActive = trimmedApiKey !== ''
    const customKeyManaged = customKeyActive && configuredKeys.includes(trimmedApiKey)

    const models = [
        // Qwen3.6 系列
        { id: 'qwen3.6-plus', name: 'qwen3.6-plus', icon: 'Sparkles', desc: 'Qwen3.6 旗舰版，最强性能', color: 'text-purple-600' },
        { id: 'qwen3.6-plus-2026-04-02', name: 'qwen3.6-plus-2026-04-02', icon: 'Calendar', desc: 'Qwen3.6 快照版', color: 'text-purple-500' },
        
        // Qwen3.5 系列
        { id: 'qwen3.5-max', name: 'qwen3.5-max', icon: 'Crown', desc: 'Qwen3.5 最强版本', color: 'text-indigo-600' },
        { id: 'qwen3.5-plus', name: 'qwen3.5-plus', icon: 'Sparkles', desc: 'Qwen3.5 均衡版', color: 'text-indigo-500' },
        { id: 'qwen3.5-plus-2026-02-15', name: 'qwen3.5-plus-2026-02-15', icon: 'Calendar', desc: 'Qwen3.5 快照版', color: 'text-indigo-400' },
        { id: 'qwen3.5-flash', name: 'qwen3.5-flash', icon: 'Zap', desc: 'Qwen3.5 轻量快速版', color: 'text-emerald-500' },
        { id: 'qwen3.5-flash-2026-02-15', name: 'qwen3.5-flash-2026-02-15', icon: 'Calendar', desc: 'Qwen3.5 Flash 快照版', color: 'text-emerald-400' },
        
        // Qwen3 系列
        { id: 'qwen3-max', name: 'qwen3-max', icon: 'Crown', desc: 'Qwen3 最强版本', color: 'text-blue-600' },
        { id: 'qwen3-235b-a22b', name: 'qwen3-235b-a22b', icon: 'Cpu', desc: 'Qwen3 MoE 旗舰模型', color: 'text-blue-500' },
        { id: 'qwen3-32b', name: 'qwen3-32b', icon: 'Box', desc: 'Qwen3 稠密版，企业部署首选', color: 'text-cyan-500' },
        
        // Qwen 经典系列
        { id: 'qwen-plus', name: 'qwen-plus', icon: 'MessageSquare', desc: 'Qwen 均衡版，性价比高', color: 'text-blue-500' },
        { id: 'qwen-plus-latest', name: 'qwen-plus-latest', icon: 'Globe', desc: 'Qwen 最新版', color: 'text-blue-400' },
        { id: 'qwen-plus-2024-12-20', name: 'qwen-plus-2024-12-20', icon: 'Calendar', desc: 'Qwen 快照版', color: 'text-blue-300' },
        { id: 'qwen-max', name: 'qwen-max', icon: 'Crown', desc: 'Qwen 旗舰版', color: 'text-blue-600' },
        { id: 'qwen-turbo', name: 'qwen-turbo', icon: 'Zap', desc: 'Qwen 轻量快速版', color: 'text-green-500' },
        { id: 'qwen-long', name: 'qwen-long', icon: 'FileText', desc: 'Qwen 长文本版，支持 1M 上下文', color: 'text-orange-500' },
        
        // Qwen-VL 多模态系列
        { id: 'qwen-vl-max', name: 'qwen-vl-max', icon: 'Image', desc: 'Qwen 视觉旗舰版', color: 'text-pink-600' },
        { id: 'qwen-vl-plus', name: 'qwen-vl-plus', icon: 'Image', desc: 'Qwen 视觉均衡版', color: 'text-pink-500' },
        { id: 'qwen-vl-v1', name: 'qwen-vl-v1', icon: 'Image', desc: 'Qwen 视觉初代版', color: 'text-pink-400' },
        
        // Qwen-MT 翻译系列
        { id: 'qwen-mt-plus', name: 'qwen-mt-plus', icon: 'Languages', desc: 'Qwen 翻译均衡版', color: 'text-teal-500' },
        { id: 'qwen-mt-turbo', name: 'qwen-mt-turbo', icon: 'Languages', desc: 'Qwen 翻译快速版', color: 'text-teal-400' },
        
        // Qwen-Coder 代码生成系列
        { id: 'qwen3-coder-next', name: 'qwen3-coder-next', icon: 'Code', desc: 'Qwen3 代码生成推荐版', color: 'text-violet-600' },
        { id: 'qwen3-coder-plus', name: 'qwen3-coder-plus', icon: 'Code', desc: 'Qwen3 代码生成增强版', color: 'text-violet-500' },
        { id: 'qwen3-coder-flash', name: 'qwen3-coder-flash', icon: 'Code', desc: 'Qwen3 代码生成快速版', color: 'text-violet-400' },
        { id: 'qwen3-coder-480B', name: 'qwen3-coder-480B', icon: 'Code', desc: 'Qwen3 代码生成旗舰版', color: 'text-violet-700' },
        { id: 'qwen-coder-turbo', name: 'qwen-coder-turbo', icon: 'Code', desc: 'Qwen 代码生成快速版', color: 'text-indigo-500' },
        
        // DeepSeek 系列
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
                accounts={displayAccounts}
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
