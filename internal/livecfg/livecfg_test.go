package livecfg

import (
	"testing"
	"time"
)

func TestModelsForToken(t *testing.T) {
	s := Snapshot{
		APIKey: "admin",
		ClientKeys: []ClientKey{
			{Name: "glm-only", Key: "sk-glm", Models: []string{"cn:glm-5.2"}},
			{Name: "free", Key: "sk-free"},
		},
	}
	if got := s.ModelsForToken("sk-glm"); len(got) != 1 || got[0] != "cn:glm-5.2" {
		t.Errorf("ModelsForToken(sk-glm) = %v", got)
	}
	if got := s.ModelsForToken("sk-free"); got != nil {
		t.Errorf("key without whitelist must be unrestricted, got %v", got)
	}
	if got := s.ModelsForToken("admin"); got != nil {
		t.Errorf("admin key must be unrestricted, got %v", got)
	}
	if got := s.ModelsForToken(""); got != nil {
		t.Errorf("empty token must be unrestricted, got %v", got)
	}
	if got := s.ModelsForToken("sk-nope"); got != nil {
		t.Errorf("unknown token must be unrestricted, got %v", got)
	}
}

func TestValidKeysFiltersExpired(t *testing.T) {
	now := time.Now()
	s := Snapshot{
		APIKey: "admin",
		ClientKeys: []ClientKey{
			{Name: "forever", Key: "sk-forever"},
			{Name: "live", Key: "sk-live", ExpiresAt: now.Add(time.Hour).Unix()},
			{Name: "expired", Key: "sk-expired", ExpiresAt: now.Add(-time.Second).Unix()},
		},
	}
	got := s.ValidKeys()
	if len(got) != 2 {
		t.Fatalf("ValidKeys len = %d, want 2 (%v)", len(got), got)
	}
	for _, k := range got {
		if k == "sk-expired" {
			t.Errorf("expired key must not be in ValidKeys")
		}
	}
	// 全量 Keys() 不过期过滤（面板展示用）。
	if len(s.Keys()) != 3 {
		t.Errorf("Keys() must keep all 3, got %d", len(s.Keys()))
	}
}

func TestExpiredBoundary(t *testing.T) {
	now := time.Now()
	// 到期时刻那一刻即失效（>= 判定），差 1 秒仍有效。
	if !(ClientKey{ExpiresAt: now.Unix()}).Expired(now) {
		t.Errorf("key expiring exactly now must be expired")
	}
	if (ClientKey{ExpiresAt: now.Add(time.Second).Unix()}).Expired(now) {
		t.Errorf("key expiring in 1s must not be expired")
	}
	if (ClientKey{}).Expired(now) {
		t.Errorf(" ExpiresAt=0 (permanent) must never expire")
	}
}
