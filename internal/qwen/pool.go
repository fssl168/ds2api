package qwen

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"ds2api/internal/config"
)

const (
	defaultMaxInflightPerTicket = 2
	defaultCooldownSeconds      = 30
	maxFailureCountBeforeCooldown = 3
)

type QwenPoolEntry struct {
	Label     string
	Ticket    string
	FailCount int
	CooldownUntil time.Time
}

type QwenPool struct {
	store                  *config.Store
	mu                     sync.RWMutex
	entries                []*QwenPoolEntry
	inUse                  map[string]int
	waiters                []chan struct{}
	maxInflightPerTicket   int
	globalMaxInflight      int
	maxQueueSize           int
	recommendedConcurrency int
}

func NewQwenPool(store *config.Store) *QwenPool {
	p := &QwenPool{
		store:                store,
		inUse:                map[string]int{},
		maxInflightPerTicket: defaultMaxInflightPerTicket,
	}
	p.Reset()
	return p
}

func (p *QwenPool) Reset() {
	accounts := p.store.Snapshot().QwenAccounts
	sort.SliceStable(accounts, func(i, j int) bool {
		iHas := accounts[i].Ticket != ""
		jHas := accounts[j].Ticket != ""
		if iHas == jHas {
			return strings.Compare(accounts[i].Label, accounts[j].Label) < 0
		}
		return iHas
	})
	entries := make([]*QwenPoolEntry, 0, len(accounts))
	for _, qa := range accounts {
		if strings.TrimSpace(qa.Ticket) == "" {
			continue
		}
		entry := &QwenPoolEntry{
			Label:  qa.Label,
			Ticket: qa.Ticket,
		}
		if existingEntry := p.findExistingEntry(qa.Label); existingEntry != nil {
			entry.FailCount = existingEntry.FailCount
			entry.CooldownUntil = existingEntry.CooldownUntil
		}
		entries = append(entries, entry)
	}
	recommended := defaultRecommendedQwenConcurrency(len(entries), p.maxInflightPerTicket)
	queueLimit := recommended * 2
	if queueLimit < 4 {
		queueLimit = 4
	}
	globalLimit := recommended
	p.mu.Lock()
	defer p.mu.Unlock()
	p.drainWaitersLocked()
	p.entries = entries
	p.inUse = map[string]int{}
	p.recommendedConcurrency = recommended
	p.maxQueueSize = queueLimit
	p.globalMaxInflight = globalLimit
	config.Logger.Info(
		"[qwen-pool] initialized",
		"total", len(entries),
		"max_inflight_per_ticket", p.maxInflightPerTicket,
		"global_max_inflight", p.globalMaxInflight,
		"recommended_concurrency", p.recommendedConcurrency,
		"max_queue_size", p.maxQueueSize,
	)
}

func (p *QwenPool) findExistingEntry(label string) *QwenPoolEntry {
	for _, e := range p.entries {
		if e.Label == label {
			return e
		}
	}
	return nil
}

func (p *QwenPool) Acquire(ctx context.Context) (*QwenPoolEntry, error) {
	p.mu.Lock()
	if entry, ok := p.acquireLocked(); ok {
		p.mu.Unlock()
		return entry, nil
	}
	if !p.canQueueLocked() {
		p.mu.Unlock()
		return nil, ErrQwenPoolExhausted
	}
	waiter := make(chan struct{})
	p.waiters = append(p.waiters, waiter)
	p.mu.Unlock()
	select {
	case <-ctx.Done():
		p.mu.Lock()
		p.removeWaiterLocked(waiter)
		p.mu.Unlock()
		return nil, ctx.Err()
	case <-waiter:
		p.mu.Lock()
		entry, ok := p.acquireLocked()
		p.mu.Unlock()
		if !ok {
			return nil, ErrQwenPoolExhausted
		}
		return entry, nil
	}
}

func (p *QwenPool) acquireLocked() (*QwenPoolEntry, bool) {
	now := time.Now()
	for _, entry := range p.entries {
		if entry.CooldownUntil.After(now) {
			continue
		}
		label := entry.Label
		if !p.canAcquireEntryLocked(entry) {
			continue
		}
		p.inUse[label]++
		p.bumpQueue(label)
		return entry, true
	}
	return nil, false
}

func (p *QwenPool) AcquireNoWait() (*QwenPoolEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry, ok := p.acquireLocked()
	if !ok {
		return nil, ErrQwenPoolExhausted
	}
	return entry, nil
}

func (p *QwenPool) Release(label string) {
	if label == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	count := p.inUse[label]
	if count <= 0 {
		return
	}
	if count == 1 {
		delete(p.inUse, label)
		p.notifyWaiterLocked()
		return
	}
	p.inUse[label] = count - 1
	p.notifyWaiterLocked()
}

func (p *QwenPool) MarkSuccess(label string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, e := range p.entries {
		if e.Label == label {
			if e.FailCount > 0 {
				e.FailCount--
			}
			break
		}
	}
}

func (p *QwenPool) MarkFailed(label string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, e := range p.entries {
		if e.Label == label {
			e.FailCount++
			if e.FailCount >= maxFailureCountBeforeCooldown {
				cooldown := time.Duration(defaultCooldownSeconds+e.FailCount*10) * time.Second
				e.CooldownUntil = time.Now().Add(cooldown)
				config.Logger.Warn(
					"[qwen-pool] entry cooldown",
					"label", label,
					"fail_count", e.FailCount,
					"cooldown_seconds", int(cooldown.Seconds()),
				)
			}
			break
		}
	}
}

func (p *QwenPool) Status() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	available := make([]string, 0, len(p.entries))
	inUseEntries := make([]string, 0, len(p.inUse))
	inUseSlots := 0
	for _, e := range p.entries {
		if e.CooldownUntil.Before(now) && p.inUse[e.Label] < p.maxInflightPerTicket {
			available = append(available, e.Label)
		}
	}
	for label, count := range p.inUse {
		if count > 0 {
			inUseEntries = append(inUseEntries, label)
			inUseSlots += count
		}
	}
	sort.Strings(inUseEntries)
	cooldownCount := 0
	for _, e := range p.entries {
		if e.CooldownUntil.After(now) {
			cooldownCount++
		}
	}
	return map[string]any{
		"available":                 len(available),
		"in_use":                    inUseSlots,
		"total":                     len(p.entries),
		"available_accounts":        available,
		"in_use_accounts":           inUseEntries,
		"cooldown_accounts":         cooldownCount,
		"max_inflight_per_ticket":   p.maxInflightPerTicket,
		"global_max_inflight":       p.globalMaxInflight,
		"recommended_concurrency":   p.recommendedConcurrency,
		"waiting":                   len(p.waiters),
		"max_queue_size":            p.maxQueueSize,
	}
}

func (p *QwenPool) SetLimits(maxInflightPerTicket, maxQueueSize, globalMaxInflight int) {
	if maxInflightPerTicket <= 0 {
		maxInflightPerTicket = defaultMaxInflightPerTicket
	}
	if maxQueueSize < 0 {
		maxQueueSize = 0
	}
	if globalMaxInflight <= 0 {
		globalMaxInflight = maxInflightPerTicket * len(p.entries)
		if globalMaxInflight <= 0 {
			globalMaxInflight = maxInflightPerTicket
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxInflightPerTicket = maxInflightPerTicket
	p.maxQueueSize = maxQueueSize
	p.globalMaxInflight = globalMaxInflight
	p.recommendedConcurrency = defaultRecommendedQwenConcurrency(len(p.entries), p.maxInflightPerTicket)
	p.notifyWaiterLocked()
	config.Logger.Info(
		"[qwen-pool] limits updated",
		"max_inflight_per_ticket", p.maxInflightPerTicket,
		"global_max_inflight", p.globalMaxInflight,
		"max_queue_size", p.maxQueueSize,
	)
}

func (p *QwenPool) canAcquireEntryLocked(entry *QwenPoolEntry) bool {
	if entry == nil || entry.Label == "" {
		return false
	}
	if time.Now().Before(entry.CooldownUntil) {
		return false
	}
	if p.inUse[entry.Label] >= p.maxInflightPerTicket {
		return false
	}
	if p.globalMaxInflight > 0 && p.currentInUseLocked() >= p.globalMaxInflight {
		return false
	}
	return true
}

func (p *QwenPool) currentInUseLocked() int {
	total := 0
	for _, n := range p.inUse {
		total += n
	}
	return total
}

func (p *QwenPool) canQueueLocked() bool {
	if p.maxQueueSize <= 0 {
		return false
	}
	return len(p.waiters) < p.maxQueueSize
}

func (p *QwenPool) notifyWaiterLocked() {
	if len(p.waiters) == 0 {
		return
	}
	waiter := p.waiters[0]
	p.waiters = p.waiters[1:]
	close(waiter)
}

func (p *QwenPool) removeWaiterLocked(waiter chan struct{}) bool {
	for i, w := range p.waiters {
		if w != waiter {
			continue
		}
		p.waiters = append(p.waiters[:i], p.waiters[i+1:]...)
		return true
	}
	return false
}

func (p *QwenPool) drainWaitersLocked() {
	for _, waiter := range p.waiters {
		close(waiter)
	}
	p.waiters = nil
}

func (p *QwenPool) bumpQueue(label string) {
	for i, e := range p.entries {
		if e.Label != label {
			continue
		}
		entry := p.entries[i]
		p.entries = append(p.entries[:i], p.entries[i+1:]...)
		p.entries = append(p.entries, entry)
		return
	}
}

func (p *QwenPool) RandomEntry() *QwenPoolEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	now := time.Now()
	var available []*QwenPoolEntry
	for _, e := range p.entries {
		if e.CooldownUntil.Before(now) {
			available = append(available, e)
		}
	}
	if len(available) == 0 {
		return nil
	}
	return available[rand.Intn(len(available))]
}

func defaultRecommendedQwenConcurrency(entryCount, maxInflight int) int {
	if entryCount <= 0 {
		return 0
	}
	if maxInflight <= 0 {
		maxInflight = defaultMaxInflightPerTicket
	}
	return entryCount * maxInflight
}

var ErrQwenPoolExhausted = func() error { return &qwenPoolError{msg: "qwen pool exhausted: no available tickets or queue full"} }()

type qwenPoolError struct{ msg string }

func (e *qwenPoolError) Error() string { return e.msg }
func (e *qwenPoolError) Is(target error) bool {
	if target == ErrQwenPoolExhausted {
		return true
	}
	return false
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
