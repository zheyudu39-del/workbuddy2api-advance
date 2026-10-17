// Package livecfg 运行期可变配置的并发安全持有者。
//
// 背景：进程启动时读入的配置是普通字段（读多写零），但管理面板允许在线改配置，
// 于是少量"可热生效"的字段需要有并发安全的读写点。此处用不可变快照 + atomic 指针：
// 读方 Load 拿到一致视图，写方 Store 整体替换，无锁无数据竞争。
//
// 只承载**读路径深、热改需求强**的少数字段；池参数/排程参数等各有既有 setter
// （pool.SetBreaker、scheduler.Reconfigure 等），不重复收编到这里。
package livecfg

import (
	"crypto/subtle"
	"sync/atomic"
	"time"
)

// Snapshot 一次读取的不可变配置视图。
type Snapshot struct {
	APIKey               string        // 网关/面板共同鉴权密钥；空 = 不鉴权
	ClientKeys           []ClientKey   // 仅可调用 /v1 的附加密钥（不能登录面板）；APIKey 为空时不生效
	SoftCooldown         time.Duration // 429 软冷却基数（<=0 时调用方回退内置默认）
	SanitizeFingerprints bool          // 出站请求体指纹脱敏
	RecordClientInfo     bool          // 请求日志是否记录调用来源（客户端 IP / UA）
}

// ClientKey 一把客户端密钥（面板「客户端密钥」卡片增删；落盘于 config.json client_keys）。
//
// 上限常量面板（创建入口）与配置加载（normalizeClientKeys）共用，口径只有一份：
const (
	MaxClientKeyHours  = 8760 // 有效期上限（小时，1 年）；0 = 永久
	MaxClientKeyModels = 256  // 单密钥模型白名单数量上限
)

type ClientKey struct {
	Name      string `json:"name"`
	Key       string `json:"key"`
	CreatedAt int64  `json:"created_at"` // unix 秒
	ExpiresAt int64  `json:"expires_at"` // unix 秒；0 = 永久有效，到点后自动失效
	Models    []string `json:"models"`  // 模型白名单（调用值即 /v1/models 的 id，如 "cn:glm-5.2"）；空 = 不限制
}

// Expired 判断密钥在 now 时刻是否已过期（ExpiresAt 为 0 表示永久，永不过期）。
func (k ClientKey) Expired(now time.Time) bool {
	return k.ExpiresAt > 0 && !now.Before(time.Unix(k.ExpiresAt, 0))
}

// ModelsForToken 返回 token 对应的客户端密钥的模型白名单。
// token 不是任何客户端密钥（主 api_key / 未命中）或该密钥未设白名单时返回 nil = 不限制。
// 过期判定由调用方（网关鉴权）先行完成，此处只做密钥→白名单映射。
func (s Snapshot) ModelsForToken(token string) []string {
	if token == "" {
		return nil
	}
	for _, k := range s.ClientKeys {
		if subtle.ConstantTimeCompare([]byte(k.Key), []byte(token)) == 1 {
			if len(k.Models) == 0 {
				return nil
			}
			return k.Models
		}
	}
	return nil
}

// Keys 返回全部客户端密钥原文（含已过期的；仅面板展示等非鉴权场景用）。
func (s Snapshot) Keys() []string {
	out := make([]string, 0, len(s.ClientKeys))
	for _, k := range s.ClientKeys {
		out = append(out, k.Key)
	}
	return out
}

// ValidKeys 返回当前仍有效的客户端密钥原文（过滤已过期；网关 /v1/* 鉴权用）。
func (s Snapshot) ValidKeys() []string {
	now := time.Now()
	out := make([]string, 0, len(s.ClientKeys))
	for _, k := range s.ClientKeys {
		if !k.Expired(now) {
			out = append(out, k.Key)
		}
	}
	return out
}

// Holder 原子持有当前快照。
type Holder struct {
	p atomic.Pointer[Snapshot]
}

// New 以初始快照构建。
func New(s Snapshot) *Holder {
	h := &Holder{}
	h.Store(s)
	return h
}

// Load 返回当前快照（Holder 为 nil 或从未 Store 时返回零值快照，调用方无需判空）。
func (h *Holder) Load() Snapshot {
	if h == nil {
		return Snapshot{}
	}
	if s := h.p.Load(); s != nil {
		return *s
	}
	return Snapshot{}
}

// Store 整体替换快照。
func (h *Holder) Store(s Snapshot) { h.p.Store(&s) }
