// clientkeys.go 客户端密钥增删：只能调用 /v1/* 的附加密钥，与面板管理密钥 api_key 分离。
// 落盘走 SaveConfig（校验 → 原子写 config.json → 热应用 livecfg），删除后下一个请求即失效。
package panel

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/linguo2625469/workbuddy2api-panel/internal/livecfg"
)

func (p *Panel) listClientKeys(w http.ResponseWriter, r *http.Request) {
	live := p.cfg.Live.Load()
	keys := live.ClientKeys
	if keys == nil {
		keys = []livecfg.ClientKey{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"keys":          keys,
		"auth_required": live.APIKey != "",
	})
}

// normalizeModelList 清洗面板提交的模型白名单：trim、丢空、去重、限量。
// 返回 (清洗后列表, 错误消息)；错误消息为空串即通过。
// 不校验模型是否真实存在——白名单按调用值（/v1/models 的 id）原样保存，
// 上游已下线的模型自然serving不到，留着也无害。
func normalizeModelList(in []string) ([]string, string) {
	if len(in) > livecfg.MaxClientKeyModels {
		return nil, fmt.Sprintf("模型白名单最多 %d 条", livecfg.MaxClientKeyModels)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, m := range in {
		m = strings.TrimSpace(m)
		if m == "" || utf8.RuneCountInString(m) > 128 || seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	return out, ""
}

func (p *Panel) createClientKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name           string   `json:"name"`
		ExpiresInHours float64  `json:"expires_in_hours"` // 0/缺省 = 永久
		Models         []string `json:"models"`           // 模型白名单（调用值即 /v1/models 的 id）；空 = 不限制
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "parse body: "+err.Error())
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || utf8.RuneCountInString(name) > 64 {
		writeErr(w, http.StatusBadRequest, "名称必填且不超过 64 字")
		return
	}
	hours := body.ExpiresInHours
	if hours < 0 || hours > livecfg.MaxClientKeyHours {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("有效期须为 0~%d 小时（0 = 永久）", livecfg.MaxClientKeyHours))
		return
	}
	models, errMsg := normalizeModelList(body.Models)
	if errMsg != "" {
		writeErr(w, http.StatusBadRequest, errMsg)
		return
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		writeErr(w, http.StatusInternalServerError, "gen key: "+err.Error())
		return
	}
	now := time.Now()
	nk := livecfg.ClientKey{Name: name, Key: "sk-" + hex.EncodeToString(raw), CreatedAt: now.Unix(), Models: models}
	if hours > 0 {
		nk.ExpiresAt = now.Add(time.Duration(hours * float64(time.Hour))).Unix()
	}

	p.clientKeyMu.Lock()
	defer p.clientKeyMu.Unlock()
	live := p.cfg.Live.Load()
	if live.APIKey == "" {
		writeErr(w, http.StatusConflict, "api_key 为空（网关未启用鉴权），请先设置 api_key")
		return
	}
	keys := append(append([]livecfg.ClientKey{}, live.ClientKeys...), nk)
	if !p.storeClientKeys(w, keys) {
		return
	}
	log.Printf("panel: 新建客户端密钥 %q（%s，白名单 %d 个模型）", name, ckExpiryLabel(nk, now), len(nk.Models))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "key": nk})
}

// ckExpiryLabel 日志用的有效期描述。
func ckExpiryLabel(k livecfg.ClientKey, now time.Time) string {
	if k.ExpiresAt == 0 {
		return "永久"
	}
	d := time.Duration(k.ExpiresAt-now.Unix()) * time.Second
	if d <= 0 {
		return "已过期"
	}
	return fmt.Sprintf("%.1f 小时后过期", d.Hours())
}

func (p *Panel) deleteClientKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err != nil || body.Key == "" {
		writeErr(w, http.StatusBadRequest, "缺少 key")
		return
	}
	p.clientKeyMu.Lock()
	defer p.clientKeyMu.Unlock()
	cur := p.cfg.Live.Load().ClientKeys
	keys := make([]livecfg.ClientKey, 0, len(cur))
	var removed *livecfg.ClientKey
	for i := range cur {
		if cur[i].Key == body.Key {
			removed = &cur[i]
			continue
		}
		keys = append(keys, cur[i])
	}
	if removed == nil {
		writeErr(w, http.StatusNotFound, "密钥不存在")
		return
	}
	if !p.storeClientKeys(w, keys) {
		return
	}
	log.Printf("panel: 删除客户端密钥 %q", removed.Name)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// storeClientKeys 以整表替换 client_keys 落盘并热应用；失败已写错误响应。
func (p *Panel) storeClientKeys(w http.ResponseWriter, keys []livecfg.ClientKey) bool {
	if p.cfg.SaveConfig == nil || p.cfg.Live == nil {
		writeErr(w, http.StatusNotImplemented, "config api not available")
		return false
	}
	raw, err := json.Marshal(map[string]any{"client_keys": keys})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return false
	}
	if _, err := p.cfg.SaveConfig(raw); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}
