package qwen

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"ds2api/internal/config"
)

func (c *Client) buildCookie(ticket string) string {
	hash := c.ticketHash(ticket)
	return fmt.Sprintf(
		"theme-mode=light; _samesite_flag_=true; tongyi_sso_ticket=%s; tongyi_sso_ticket_hash=%s; XSRF-TOKEN=%s",
		ticket, hash, c.xsrf,
	)
}

func (c *Client) pickTicket() string {
	if entry := c.pool.RandomEntry(); entry != nil {
		return entry.Ticket
	}
	return ""
}

func (c *Client) pickEntry() *QwenPoolEntry {
	return c.pool.RandomEntry()
}

func (c *Client) generateNonce() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func (c *Client) calibrateTime(ctx context.Context) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, QwenSecCalibURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return time.Now().UnixMilli(), nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			MillisTimeStamp string `json:"millisTimeStamp"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)
	if result.Data.MillisTimeStamp != "" {
		var ts int64
		fmt.Sscanf(result.Data.MillisTimeStamp, "%d", &ts)
		return ts, nil
	}
	return time.Now().UnixMilli(), nil
}

func (c *Client) setSecurityHeaders(r *http.Request, ticket string, chatID string, reqTimestamp int64) {
	if c.secMgr != nil && c.secMgr.isValid() {
		headers := c.secMgr.BuildChatHeaders(chatID, reqTimestamp, ticket)
		for k, v := range headers {
			r.Header.Set(k, v[0])
		}
		return
	}

	r.Header.Set("Accept", "application/json, text/event-stream, text/plain, */*")
	r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	r.Header.Set("Cache-Control", "no-cache")
	r.Header.Set("Content-Type", "application/json;charset=UTF-8")
	r.Header.Set("Origin", "https://www.qianwen.com")
	r.Header.Set("Pragma", "no-cache")
	r.Header.Set("Referer", fmt.Sprintf("https://www.qianwen.com/chat/%s", chatID))
	r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"`)
	r.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	r.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	r.Header.Set("Sec-Fetch-Dest", "empty")
	r.Header.Set("Sec-Fetch-Mode", "cors")
	r.Header.Set("Sec-Fetch-Site", "same-site")
	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	r.Header.Set("bx-v", "2.5.31")

	r.Header.Set("eo-clt-acs-ve", "1.0.0")
	r.Header.Set("eo-clt-acs-kp", c.ticketHash(ticket))

	r.Header.Set("clt-acs-caer", "vrad")
	r.Header.Set("clt-acs-request-params", "biz_id,chat_client,device,fr,pr,ut,la,tz,nonce,timestamp")
	r.Header.Set("clt-acs-sign", computeRealSign(chatID, reqTimestamp))
	r.Header.Set("clt-acs-reqt", fmt.Sprintf("%d", reqTimestamp))

	r.Header.Set("x-xsrf-token", c.xsrf)
	r.Header.Set("x-chat-id", chatID)
	r.Header.Set("x-deviceid", c.deviceID)
	r.Header.Set("x-platform", "pc_tongyi")

	r.Header.Set("Cookie", c.buildCookie(ticket))
}

func computeRealSign(chatID string, timestamp int64) string {
	key := []byte("qwen_chat_sign_key_v1")
	msg := fmt.Sprintf("%s%d%s%s", chatID, timestamp, "", "biz_id,chat_client,device,fr,pr,ut,la,tz,nonce,timestamp")
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))[:32]
}

func (c *Client) Preload(ctx context.Context) error {
	status := c.pool.Status()
	total := status["total"].(int)
	if total > 0 {
		config.Logger.Info("[qwen] preload complete", "tickets", total, "device_id", c.deviceID[:8]+"...")
	} else {
		config.Logger.Warn("[qwen] no tickets configured")
	}

	if c.secMgr != nil && total > 0 {
		if err := c.secMgr.Register(ctx); err != nil {
			config.Logger.Warn("[qwen] security register failed", "error", err)
		}
	}

	return nil
}
