package types

import (
	"context"
	"time"

	"github.com/athenianco/cloud-common/report"
)

type RepoEventType string

const (
	RepoAdded    = RepoEventType("added")
	RepoUpdated  = RepoEventType("updated")
	RepoRemoved  = RepoEventType("removed")
	RepoComplete = RepoEventType("complete")
	RepoIndexed  = RepoEventType("indexed")
	OrgRenamed   = RepoEventType("org-renamed")
)

type RenameEvent struct {
	NodeID  string `json:"node_id"`
	Name    string `json:"name"`
	NameOld string `json:"name_old,omitempty"`
}

type RepoEvent struct {
	EventID      EventID       `json:"event_id"`
	Timestamp    time.Time     `json:"ts,omitempty"`
	AccID        AccID         `json:"acc_id"`
	Type         RepoEventType `json:"type"`
	OrgName      string        `json:"org_name,omitempty"`   // set for OrgRenamed; deprecated
	OrgRename    *RenameEvent  `json:"org_rename,omitempty"` // set for OrgRenamed events
	NodeIDs      []NodeID      `json:"node_id,omitempty"`
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
