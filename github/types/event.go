package types

import (
	"context"
	"time"

	"github.com/athenianco/cloud-common/report"
)

type RepoEventType string

const (
	// OrgRenamed event triggers when the organization changes their name (GitHub "login").
	OrgRenamed = RepoEventType("org-renamed")
	// RepoAdded event triggers when a new or existing repository is added to Athenian.
	RepoAdded = RepoEventType("added")
	// RepoUpdated event triggers when a repository is renamed.
	RepoUpdated = RepoEventType("updated")
	// RepoRemoved event triggers when a repository is deleted or removed from Athenian.
	RepoRemoved = RepoEventType("removed")
	// RepoFetched event triggers when a repository is completely fetched (or fetch times out).
	RepoFetched = RepoEventType("fetched")
	// RepoComplete event triggers when a repository is completely fetched (or fetch times out).
	//
	// Deprecated: use RepoFetched instead.
	RepoComplete = RepoEventType("complete")
	// RepoIndexed event triggers when a repository is indexed in the database (VACUUM and other similar operations).
	//
	// Deprecated: use RepoFetched instead.
	RepoIndexed = RepoEventType("indexed")
)

type RenameEvent struct {
	NodeID  string  `json:"node_id,omitempty"`
	GID     GraphID `json:"gid,omitempty"`
	Name    string  `json:"name"`
	NameOld string  `json:"name_old,omitempty"`
}

type RepoEvent struct {
	EventID      EventID       `json:"event_id"`
	Timestamp    time.Time     `json:"ts,omitempty"`
	AccID        AccID         `json:"acc_id"`
	Type         RepoEventType `json:"type"`
	OrgName      string        `json:"org_name,omitempty"`   // set for OrgRenamed; deprecated
	OrgRename    *RenameEvent  `json:"org_rename,omitempty"` // set for OrgRenamed events
	NodeIDs      []NodeID      `json:"node_id,omitempty"`
	GIDs         []GraphID     `json:"gids,omitempty"`
	FullNames    []string      `json:"full_name,omitempty"`
	FullNamesOld []string      `json:"full_name_old,omitempty"` // set for RepoUpdated event
	NodesTotal   uint64        `json:"nodes_total,omitempty"`
}

func (ev *RepoEvent) EventContext() EventContext {
	var ectx EventContext
	ectx.EventID = ev.EventID
	ectx.AccountID = ev.AccID
	return ectx
}

func WithRepoEvent(ctx context.Context, ev *RepoEvent) context.Context {
	ctx = WithAccount(ctx, ev.AccID)
	ctx = WithEvent(ctx, ev.EventID)
	ctx = report.WithStringValue(ctx, "github.repos.event", string(ev.Type))
	ctx = WithNodeIDs(ctx, NodeIDsToStrings(ev.NodeIDs))
	return ctx
}
