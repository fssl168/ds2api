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
        // Qwen 系列 (与后端 qianwen.com Web API 支持的模型保持一致)
        { id: 'qwen', name: 'qwen', icon: 'MessageSquare', desc: 'Qwen 基础版', color: 'text-blue-500' },
        { id: 'qwen-max', name: 'qwen-max', icon: 'Crown', desc: 'Qwen3-Max 最强版本', color: 'text-blue-600' },
        { id: 'qwen-max-thinking', name: 'qwen-max-thinking', icon: 'Brain', desc: 'Qwen3-Max-Thinking 深度思考版', color: 'text-violet-600' },
        { id: 'qwen-plus', name: 'qwen-plus', icon: 'Sparkles', desc: 'Qwen3-Plus 均衡版', color: 'text-blue-500' },
        { id: 'qwen-coder', name: 'qwen-coder', icon: 'Code', desc: 'Qwen3-Coder 代码生成版', color: 'text-violet-500' },
        { id: 'qwen-flash', name: 'qwen-flash', icon: 'Zap', desc: 'Qwen3-Flash 轻量快速版', color: 'text-emerald-500' },
        { id: 'qwen3.5-plus', name: 'qwen3.5-plus', icon: 'Sparkles', desc: 'Qwen3.5-Plus 均衡版', color: 'text-indigo-500' },
        { id: 'qwen3.5-flash', name: 'qwen3.5-flash', icon: 'Zap', desc: 'Qwen3.5-Flash 轻量快速版', color: 'text-emerald-500' },
        { id: 'qwen3.6-plus', name: 'qwen3.6-plus', icon: 'Sparkles', desc: 'Qwen3.6-Plus 旗舰版', color: 'text-purple-600' },
        { id: 'qwen3.6-plus-2026-04-02', name: 'qwen3.6-plus-2026-04-02', icon: 'Calendar', desc: 'Qwen3.6-Plus 快照版', color: 'text-purple-500' },
        
        // DeepSeek 系列
        { id: 'deepseek-chat', name: 'deepseek-chat', icon: 'MessageSquare', desc: t('apiTester.models.chat'), color: 'text-amber-500' },
        { id: 'deepseek-reasoner', name: 'deepseek-reasoner', icon: 'Cpu', desc: t('apiTester.models.reasoner'), color: 'text-amber-600' },
        { id: 'deepseek-chat-search', name: 'deepseek-chat-search', icon: 'SearchIcon', desc: t('apiTester.models.chatSearch'), color: 'text-cyan-500' },
        { id: 'deepseek-reasoner-search', name: 'deepseek-reasoner-search', icon: 'SearchIcon', desc: t('apiTester.models.reasonerSearch'), color: 'text-cyan-600' },
        // DeepSeek Expert 系列
        { id: 'deepseek-expert-chat', name: 'deepseek-expert-chat', icon: 'Award', desc: 'DeepSeek 专家聊天版', color: 'text-purple-500' },
        { id: 'deepseek-expert-reasoner', name: 'deepseek-expert-reasoner', icon: 'Brain', desc: 'DeepSeek 专家推理版', color: 'text-purple-600' },
        { id: 'deepseek-expert-chat-search', name: 'deepseek-expert-chat-search', icon: 'SearchIcon', desc: 'DeepSeek 专家聊天搜索版', color: 'text-indigo-500' },
        { id: 'deepseek-expert-reasoner-search', name: 'deepseek-expert-reasoner-search', icon: 'SearchIcon', desc: 'DeepSeek 专家推理搜索版', color: 'text-indigo-600' },
        // DeepSeek Vision 系列
        { id: 'deepseek-vision-chat', name: 'deepseek-vision-chat', icon: 'Image', desc: 'DeepSeek 视觉聊天版', color: 'text-pink-500' },
        { id: 'deepseek-vision-reasoner', name: 'deepseek-vision-reasoner', icon: 'Image', desc: 'DeepSeek 视觉推理版', color: 'text-pink-600' },
        { id: 'deepseek-vision-chat-search', name: 'deepseek-vision-chat-search', icon: 'Image', desc: 'DeepSeek 视觉聊天搜索版', color: 'text-rose-500' },
        { id: 'deepseek-vision-reasoner-search', name: 'deepseek-vision-reasoner-search', icon: 'Image', desc: 'DeepSeek 视觉推理搜索版', color: 'text-rose-600' },
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
