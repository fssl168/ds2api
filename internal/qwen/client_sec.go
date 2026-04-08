package qwen

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"ds2api/internal/config"
)

type SecurityManager struct {
	httpClient *http.Client
	client     *Client
	mu         sync.RWMutex

	actkn    string
	dvidn    string
	sacsft   string
	snver    string
	bacsfts  []string
	expireAt time.Time
}

func newSecurityManager(client *Client) *SecurityManager {
	return &SecurityManager{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       5,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
		client: client,
	}
}

func (sm *SecurityManager) Actkn() string  { sm.mu.RLock(); defer sm.mu.RUnlock(); return sm.actkn }
func (sm *SecurityManager) Dvidn() string  { sm.mu.RLock(); defer sm.mu.RUnlock(); return sm.dvidn }
func (sm *SecurityManager) Sacsft() string { sm.mu.RLock(); defer sm.mu.RUnlock(); return sm.sacsft }
func (sm *SecurityManager) Snver() string  { sm.mu.RLock(); defer sm.mu.RUnlock(); return sm.snver }

func (sm *SecurityManager) isValid() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.actkn != "" && time.Now().Before(sm.expireAt)
}

func (sm *SecurityManager) Register(ctx context.Context) error {
	umidToken, err := sm.fetchUMID(ctx)
	if err != nil {
		config.Logger.Warn("[qwen-sec] fetch umid failed, will try register anyway", "error", err)
	} else {
		sm.client.setUMIDToken(umidToken)
	}

	chid := sm.client.chid
	ticket := sm.client.pickTicket()
	if ticket == "" {
		return fmt.Errorf("no ticket available")
	}

	browserFeatures := collectBrowserFeatures()
	fingerprint := collectFingerprint()

	body := map[string]any{
		"features":            browserFeatures,
		"fingerprint":         fingerprint,
		"businessScene":       "qwen_chat",
		"chid":                chid,
		"unifyRelateGenerate": []string{},
	}

	bodyJSON, _ := json.Marshal(body)
	registerURL := fmt.Sprintf("%s?chid=%s", QwenSecRegisterURL, chid)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("create register req: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", sm.client.buildCookie(ticket))
	setSecRegisterHeaders(req)
	req.Header.Set("bx-umidtoken", sm.client.getUMIDToken())
	req.Header.Set("bx-v", "2.5.31")
	req.Header.Set("eo-clt-acs-bx-intss", "1")

	resp, err := sm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("register http do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed status %d: %s", resp.StatusCode, truncateStr(string(respBody), 500))
	}

	var regResp struct {
		Status int    `json:"status"`
		Code   string `json:"code"`
		Msg    string `json:"msg"`
		Data   struct {
			Actkn       string   `json:"eo-clt-actkn"`
			Dvidn       string   `json:"eo-clt-dvidn"`
			Bacsfts     []string `json:"eo-clt-bacsft"`
			Snver       string   `json:"eo-clt-snver"`
			ActknDl     int64    `json:"eo-clt-actkn-dl"`
			ExpireTime  int64    `json:"expireTime"`
			UnifyRelate []struct {
				BusinessScene string   `json:"businessScene"`
				Actkn         string   `json:"eo-clt-actkn"`
				Dvidn         string   `json:"eo-clt-dvidn"`
				Sacsfts       []string `json:"eo-clt-bacsft"`
				Snver         string   `json:"eo-clt-snver"`
			} `json:"unifyRelate"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return fmt.Errorf("unmarshal refresh resp: %w (body: %s)", err, truncateStr(string(respBody), 300))
	}

	if regResp.Status != 0 || regResp.Data.Actkn == "" {
		// Provide detailed error messages similar to DeepSeek's approach
		errorMsg := fmt.Sprintf("code=%s msg=%s", regResp.Code, regResp.Msg)
		
		// Check for common authentication error patterns
		msgLower := strings.ToLower(regResp.Msg)
		if strings.Contains(msgLower, "ticket") || strings.Contains(msgLower, "invalid") {
			if strings.Contains(msgLower, "expir") {
				errorMsg = fmt.Sprintf("Qwen ticket expired: code=%s msg=%s", regResp.Code, regResp.Msg)
			} else if strings.Contains(msgLower, "wrong") || strings.Contains(msgLower, "incorrect") {
				errorMsg = fmt.Sprintf("Qwen ticket invalid: code=%s msg=%s", regResp.Code, regResp.Msg)
			} else {
				errorMsg = fmt.Sprintf("Qwen authentication failed: code=%s msg=%s", regResp.Code, regResp.Msg)
			}
		} else if strings.Contains(msgLower, "security") || strings.Contains(msgLower, "risk") {
			errorMsg = fmt.Sprintf("Qwen security validation failed: code=%s msg=%s", regResp.Code, regResp.Msg)
		}
		
		return fmt.Errorf("refresh failed: %s (body: %s)", errorMsg, truncateStr(string(respBody), 500))
	}

	d := regResp.Data
	sacsft := ""
	if len(d.Bacsfts) > 0 {
		sacsft = d.Bacsfts[0]
	}

	sm.mu.Lock()
	sm.actkn = d.Actkn
	sm.dvidn = d.Dvidn
	sm.sacsft = sacsft
	snver := d.Snver
	if snver == "" {
		snver = "lv"
	}
	sm.snver = snver
	sm.bacsfts = d.Bacsfts
	if d.ActknDl > 0 {
		sm.expireAt = time.Unix(d.ActknDl, 0)
	} else if d.ExpireTime > 0 {
		sm.expireAt = time.UnixMilli(d.ExpireTime)
	} else {
		sm.expireAt = time.Now().Add(30 * time.Minute)
	}
	sm.mu.Unlock()

	config.Logger.Info("[qwen-sec] refresh successfully",
		"actkn_len", len(d.Actkn),
		"dvidn_len", len(d.Dvidn),
		"sacsft_len", len(sacsft),
		"snver", snver,
		"bacsft_count", len(d.Bacsfts),
		"expire_in", time.Until(sm.expireAt).String(),
	)

	for _, rel := range d.UnifyRelate {
		if rel.BusinessScene != "" && rel.Actkn != "" {
			config.Logger.Info("[qwen-sec] unify relate scene",
				"scene", rel.BusinessScene,
				"actkn_len", len(rel.Actkn))
		}
	}

	return nil
}

func (sm *SecurityManager) BuildChatHeaders(chatID string, timestamp int64, ticket string) http.Header {
	h := http.Header{}

	sm.mu.RLock()
	actkn := sm.actkn
	dvidn := sm.dvidn
	sacsft := sm.sacsft
	snver := sm.snver
	kp := sm.client.ticketHash(ticket)
	sm.mu.RUnlock()

	params := "biz_id,chat_client,device,fr,pr,ut,la,tz,nonce,timestamp"

	sigInput := fmt.Sprintf("%s%d%s%s", chatID, timestamp, kp, params)
	signValue := computeHMAC256(":"+kp, sigInput)

	h.Set("Accept", "application/json, text/event-stream, text/plain, */*")
	h.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	h.Set("Cache-Control", "no-cache")
	h.Set("Content-Type", "application/json;charset=UTF-8")
	h.Set("Origin", "https://www.qianwen.com")
	h.Set("Pragma", "no-cache")
	h.Set("Referer", fmt.Sprintf("https://www.qianwen.com/chat/%s", chatID))
	h.Set("Sec-Ch-Ua", `"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"`)
	h.Set("Sec-Ch-Ua-Mobile", "?0")
	h.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-site")
	h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	h.Set("bx-v", "2.5.31")
	h.Set("bx-et", sm.client.getET())
	h.Set("bx-ua", sm.client.getUA())
	h.Set("bx-umidtoken", sm.client.getUMIDToken())

	h.Set("eo-clt-acs-ve", "1.0.0")
	h.Set("eo-clt-sacsft", sacsft)
	h.Set("eo-clt-snver", snver)
	h.Set("eo-clt-dvidn", dvidn)
	h.Set("eo-clt-actkn", actkn)
	h.Set("eo-clt-acs-kp", kp)

	h.Set("clt-acs-caer", "vrad")
	h.Set("clt-acs-request-params", params)
	h.Set("clt-acs-sign", signValue)
	h.Set("clt-acs-reqt", fmt.Sprintf("%d", timestamp))

	h.Set("x-xsrf-token", sm.client.xsrf)
	h.Set("x-chat-id", chatID)
	h.Set("x-deviceid", sm.client.deviceID)
	h.Set("x-platform", "pc_tongyi")

	cookie := sm.client.buildCookie(ticket)
	h.Set("Cookie", cookie)

	return h
}

func gzipCompress(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

func encodeBase64(data []byte) string {
	const b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder
	result.Grow(len(data)*4/3 + 4)
	val := uint(0)
	nbits := uint(0)
	for _, b := range data {
		val = val<<8 | uint(b)
		nbits += 8
		for nbits >= 6 {
			nbits -= 6
			result.WriteByte(b64[val>>nbits&0x3F])
		}
	}
	if nbits > 0 {
		val <<= 6 - nbits
		result.WriteByte(b64[val&0x3F])
	}
	for result.Len()%4 != 0 {
		result.WriteByte('=')
	}
	return result.String()
}

func computeHMAC256(key, msg string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func collectBrowserFeatures() map[string]interface{} {
	features := make(map[string]interface{})

	screenW, screenH := 1920, 1080
	features["screen_resolution"] = fmt.Sprintf("%dx%d", screenW, screenH)
	features["screen_color_depth"] = "24"
	features["language"] = "zh-CN"
	features["timezone"] = "480"
	features["session_storage"] = 1
	features["local_storage"] = 1
	features["indexed_db"] = 1
	features["open_database"] = 1
	features["cpu_class"] = "unknown"
	features["navigator_platform"] = "Win32"
	features["webgl_vendor"] = "Google Inc. (NVIDIA)"
	features["webgl_renderer"] = "ANGLE (NVIDIA, NVIDIA GeForce GTX 1060 6GB Direct3D11 vs_5_0 ps_5_0, D3D11)"

	canvasFeatures := []string{
		"canvas winding", "canvas todataurl",
	}
	features["canvas_features"] = canvasFeatures

	plugins := []string{
		"internal-pdf-viewer Chrome PDF Plugin",
		"internal-edge-pdf Microsoft Edge PDF Plugin",
	}
	features["plugins"] = plugins

	fonts := []string{"Arial", "Consolas", "Microsoft YaHei", "Monaco", "Segoe UI", "Tahoma", "Times New Roman", "Verdana"}
	features["fonts"] = fonts

	features["audio"] = "no"
	features["video"] = "no"
	features["touch_support"] = 0
	features["hardware_concurrency"] = 8
	features["device_memory"] = 8
	features["max_touch_points"] = 0

	return features
}

func collectFingerprint() map[string]interface{} {
	fp := make(map[string]interface{})
	fp["user_agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	fp["platform"] = "Win32"
	fp["vendor"] = "Google Inc."
	fp["app_version"] = "5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	return fp
}

var letterBytes = []byte("abcdefghijklmnopqrstuvwxyz")

func generateChid() string {
	b := make([]byte, 17)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	ts := time.Now().UnixMilli()
	return fmt.Sprintf("%d%s", ts%10000000000, string(b))
}

func setSecRegisterHeaders(r *http.Request) {
	r.Header.Set("Sec-Fetch-Dest", "empty")
	r.Header.Set("Sec-Fetch-Mode", "cors")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	r.Header.Set("Accept", "*/*")
	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
}

func setSecRefreshHeaders(r *http.Request, cookie string, umidToken string) {
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	r.Header.Set("Cache-Control", "no-cache")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cookie", cookie)
	r.Header.Set("Origin", "https://www.qianwen.com")
	r.Header.Set("Pragma", "no-cache")
	r.Header.Set("Referer", "https://www.qianwen.com/chat/")
	r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"`)
	r.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	r.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	r.Header.Set("Sec-Fetch-Dest", "empty")
	r.Header.Set("Sec-Fetch-Mode", "cors")
	r.Header.Set("Sec-Fetch-Site", "same-site")
	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	r.Header.Set("bx-umidtoken", umidToken)
	r.Header.Set("eo-clt-acs-bx-intss", "1")
	r.Header.Set("eo-clt-sftcnt", "100")
	r.Header.Set("clt-acs-caer", "vrad")
}

func (sm *SecurityManager) fetchUMID(ctx context.Context) (string, error) {
	fpData := buildUMIDFingerprint()
	form := url.Values{}
	form.Set("data", fpData)
	form.Set("xa", "wagbridgead-sm-nginx-quarkpc-security-calibration-time-web")
	form.Set("xt", "")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, QwenUMIDURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create umid req: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://www.qianwen.com")
	req.Header.Set("Referer", "https://www.qianwen.com/chat/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	resp, err := sm.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("umid http do: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("umid failed status %d: %s", resp.StatusCode, truncateStr(string(body), 300))
	}

	var umidResp struct {
		TN string `json:"tn"`
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &umidResp); err != nil {
		return "", fmt.Errorf("unmarshal umid resp: %w (body: %s)", err, truncateStr(string(body), 200))
	}

	if umidResp.TN == "" {
		return "", fmt.Errorf("empty umid tn (body: %s)", truncateStr(string(body), 200))
	}

	config.Logger.Info("[qwen-sec] got umid token",
		"tn_len", len(umidResp.TN),
		"id_len", len(umidResp.ID),
	)

	return umidResp.TN, nil
}

func buildUMIDFingerprint() string {
	fpParts := []string{
		"screen=" + strconv.Itoa(1920) + "*" + strconv.Itoa(1080),
		"language=zh-CN",
		"platform=Win32",
		"cpu=unknown",
		"vendor=Google Inc.",
		"webgl=Google Inc. (NVIDIA)",
		"touch=0",
		"memory=8",
		"cores=8",
		"tz=480",
		"storage=1",
		"indexedDB=1",
		"openDatabase=1",
		"canvas=1",
		"fonts=10",
		"plugins=2",
	}

	rawFP := strings.Join(fpParts, "|")
	fpBytes := []byte(rawFP)
	encoded := base64.StdEncoding.EncodeToString(fpBytes)
	return "107!" + encoded
}
