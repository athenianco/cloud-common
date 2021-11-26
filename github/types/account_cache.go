package types

import (
	"context"
	"sync"
	"time"

	"github.com/athenianco/cloud-common/dbs"
	"github.com/athenianco/cloud-common/report"
)

const (
	accCacheDuration = 10 * time.Minute
)

// NewAccountCache creates an in-memory account records cache.
func NewAccountCache(db AccountGetter) AccountGetter {
	if _, ok := db.(*accountCache); ok {
		return db
	}
	return &accountCache{
		db:     db,
		byAcc:  make(map[AccID]*accCacheItem),
		byInst: make(map[eventAccKey]*accCacheItem),
	}
}

type eventAccKey struct {
	App  AthenianAppID
	Inst InstallID
}

type accCacheItem struct {
	Loaded time.Time
	Err    error
	Inst   *Account
}

type accountCache struct {
	db AccountGetter

	mu     sync.RWMutex
	byAcc  map[AccID]*accCacheItem
	byInst map[eventAccKey]*accCacheItem
}

func (c *accountCache) GetAccountById(ctx context.Context, ictx InstallContext) (*Account, error) {
	var r *accCacheItem
	ekey := eventAccKey{App: ictx.AthenianAppID, Inst: ictx.InstallID}
	if ictx.AccountID != 0 {
		c.mu.RLock()
		r = c.byAcc[ictx.AccountID]
		c.mu.RUnlock()
	}
	if r == nil && ictx.AthenianAppID != 0 && ictx.InstallID != 0 {
		c.mu.RLock()
		r = c.byInst[ekey]
		c.mu.RUnlock()
	}
	now := time.Now()
	if r != nil && now.Sub(r.Loaded) < accCacheDuration {
		report.Info(ctx, "using cached account info: expires in %v", accCacheDuration-now.Sub(r.Loaded))
		return r.Inst, r.Err
	}
	inst, err := c.db.GetAccountById(ctx, ictx)
	if err != nil && err != dbs.ErrNotFound {
		return inst, err
	}
	r = &accCacheItem{Loaded: now, Err: err, Inst: inst}
	if inst != nil {
		ictx = inst.InstallContext
		ekey = eventAccKey{App: ictx.AthenianAppID, Inst: ictx.InstallID}
	}
	if ictx.AccountID != 0 || (ictx.AthenianAppID != 0 && ictx.InstallID != 0) {
		c.mu.Lock()
		if ictx.AccountID != 0 {
			c.byAcc[ictx.AccountID] = r
		}
		if ictx.AthenianAppID != 0 && ictx.InstallID != 0 {
			c.byInst[ekey] = r
		}
		c.mu.Unlock()
	}
	return inst, err
}
