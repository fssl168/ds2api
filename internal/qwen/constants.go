package qwen

const (
	QwenHost           = "chat2.qianwen.com"
	QwenChatURL        = "https://" + QwenHost + "/api/v2/chat"
	QwenSecCalibURL    = "https://sec.qianwen.com/api/calibration/getMillisTimeStamp"
	QwenSecRegisterURL = "https://sec.qianwen.com/security/external/access/register"
	QwenSecRefreshURL  = "https://sec.qianwen.com/security/external/access/refresh"
	QwenUMIDURL        = "https://ynuf.aliapp.org/service/um.json"

	MaxRetryCount = 3
	RetryDelayMs  = 3000
)
