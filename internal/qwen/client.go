package qwen

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"ds2api/internal/config"
)

type Client struct {
	httpClient *http.Client
	pool       *QwenPool
	mu         sync.RWMutex

	deviceID  string
	chid      string
	xsrf      string
	secMgr    *SecurityManager
	umidToken string
}

func NewClient(store *config.Store) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		},
		pool:     NewQwenPool(store),
		deviceID: uuid.New().String(),
		chid:     generateChid(),
		xsrf:     uuid.New().String(),
	}
	c.secMgr = newSecurityManager(c)
	return c
}

func (c *Client) SetTickets(tickets []string) {
	c.pool.Reset()
}

func (c *Client) TicketCount() int {
	return c.pool.Status()["total"].(int)
}

func (c *Client) Tickets() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entries := c.pool.entries
	result := make([]string, 0, len(entries))
	for _, e := range entries {
		result = append(result, e.Ticket)
	}
	return result
}

func (c *Client) Pool() *QwenPool {
	return c.pool
}

func (c *Client) ResetTickets() {
	c.pool.Reset()
}

func (c *Client) DeviceID() string {
	return c.deviceID
}

func (c *Client) ticketHash(ticket string) string {
	h := sha256.Sum256([]byte(ticket))
	return fmt.Sprintf("tytk_hash:%x", h[:16])
}

func (c *Client) getET() string {
	return generateFakeET()
}

func (c *Client) getUA() string {
	return generateFakeUA()
}

func (c *Client) getUMIDToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.umidToken != "" {
		return c.umidToken
	}
	return generateFakeUMIDToken()
}

func (c *Client) setUMIDToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.umidToken = token
}

func (c *Client) SecurityManager() *SecurityManager {
	return c.secMgr
}

func generateFakeET() string {
	return generateBxET()
}

func generateFakeUA() string {
	return generateBxUA()
}

func generateBxET() string {
	payload := make([]byte, 368)
	rand.Read(payload)
	ts := time.Now().UnixMilli()
	tsBytes := []byte(fmt.Sprintf("%d", ts))
	combined := append(payload, tsBytes...)
	h := hmac.New(sha256.New, []byte("baxia_et_salt"))
	h.Write(combined)
	sig := h.Sum(nil)
	full := append(combined, sig...)
	encoded := base64.RawURLEncoding.EncodeToString(full)
	if len(encoded) > 492 {
		encoded = encoded[:492]
	}
	for len(encoded) < 490 {
		n, _ := rand.Int(rand.Reader, big.NewInt(256))
		encoded += base64.RawURLEncoding.EncodeToString([]byte{byte(n.Int64())})
	}
	return encoded + "."
}

func generateBxUA() string {
	fp := buildBrowserFingerprint()
	fpJSON, _ := json.Marshal(fp)
	compressed := gzipCompressForUA(fpJSON)
	encoded := base64.StdEncoding.EncodeToString(compressed)
	return "231!" + encoded
}

type browserFingerprint struct {
	Screen    screenInfo    `json:"s"`
	Navigator navigatorInfo `json:"n"`
	Hardware  hardwareInfo  `json:"h"`
	WebGL     webglInfo     `json:"g"`
	Canvas    canvasInfo    `json:"c"`
	Fonts     []string      `json:"f"`
	Plugins   []pluginInfo  `json:"p"`
	Storage   storageInfo   `json:"st"`
	Timestamp int64         `json:"t"`
	Entropy   string        `json:"e"`
}

type screenInfo struct {
	Width  int `json:"w"`
	Height int `json:"h"`
	AvailW int `json:"aw"`
	AvailH int `json:"ah"`
	ColorD int `json:"cd"`
	PixelD int `json:"pd"`
}

type navigatorInfo struct {
	Platform   string `json:"pl"`
	Language   string `json:"la"`
	Languages  string `json:"las"`
	UserAgent  string `json:"ua"`
	Cookie     bool   `json:"ck"`
	DoNotTrack string `json:"dnt"`
}

type hardwareInfo struct {
	Cores     int    `json:"cores"`
	Memory    int    `json:"mem"`
	DeviceMem int    `json:"dm"`
	MaxTouch  int    `json:"touch"`
	Vendor    string `json:"vendor"`
}

type webglInfo struct {
	Vendor   string `json:"v"`
	Renderer string `json:"r"`
}

type canvasInfo struct {
	HashW string `json:"hw"`
	HashH string `json:"hh"`
}

type pluginInfo struct {
	Name    string `json:"n"`
	Version string `json:"v"`
}

type storageInfo struct {
	IndexedDB      bool `json:"idb"`
	LocalStorage   bool `json:"ls"`
	SessionStorage bool `json:"ss"`
	OpenDB         bool `json:"odb"`
}

func buildBrowserFingerprint() browserFingerprint {
	fontList := []string{
		"Arial", "Arial Black", "Comic Sans MS", "Courier New", "Georgia",
		"Impact", "Lucida Console", "Lucida Sans Unicode", "Palatino Linotype",
		"Tahoma", "Times New Roman", "Trebuchet MS", "Verdana",
		"Microsoft YaHei", "SimHei", "SimSun", "KaiTi", "FangSong",
	}
	plugins := []pluginInfo{
		{Name: "PDF Viewer", Version: "Chrome PDF Plugin"},
		{Name: "Chrome PDF Viewer", Version: ""},
		{Name: "Native Client", Version: ""},
	}
	entropyRaw := make([]byte, 32)
	rand.Read(entropyRaw)
	entropy := fmt.Sprintf("%x", entropyRaw)

	return browserFingerprint{
		Screen: screenInfo{
			Width: 1920, Height: 1080, AvailW: 1920, AvailH: 1040,
			ColorD: 24, PixelD: 1,
		},
		Navigator: navigatorInfo{
			Platform:   "Win32",
			Language:   "zh-CN",
			Languages:  "zh-CN,zh,en-US,en",
			UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36",
			Cookie:     true,
			DoNotTrack: "null",
		},
		Hardware: hardwareInfo{
			Cores: 8, Memory: 8, DeviceMem: 8, MaxTouch: 0,
			Vendor: "Google Inc.",
		},
		WebGL: webglInfo{
			Vendor:   "Google Inc. (NVIDIA)",
			Renderer: "ANGLE (NVIDIA, NVIDIA GeForce GTX 1060 6GB Direct3D11 vs_5_0 ps_5_0)",
		},
		Canvas: canvasInfo{
			HashW: "a1b2c3d4e5f6",
			HashH: "7890abcdef1234",
		},
		Fonts:   fontList,
		Plugins: plugins,
		Storage: storageInfo{
			IndexedDB: true, LocalStorage: true, SessionStorage: true, OpenDB: true,
		},
		Timestamp: time.Now().UnixMilli(),
		Entropy:   entropy,
	}
}

func gzipCompressForUA(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

func generateFakeUMIDToken() string {
	raw := make([]byte, 44)
	cryptoRandRead(raw)
	return base64.StdEncoding.EncodeToString(raw)
}

func cryptoRandRead(b []byte) {
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			n, _ := rand.Int(rand.Reader, big.NewInt(256))
			b[i] = byte(n.Int64())
		}
	}
}
