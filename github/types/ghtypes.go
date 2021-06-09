package types

import (
	"context"
	"errors"
	"strconv"

	"github.com/athenianco/cloud-common/report"
)

var ErrNoAppMeta = errors.New("no application metadata")

// AccID is a Athenian Github Account ID.
// It combines a Github AppID with a Github InstallID and can uniquely identify an installation.
type AccID int64

// AthenianAppID is a Athenian Github App ID.
// It is different from AppID, because Github's AppID is unique only in the context of a single Github instance.
// In means that AppIDs may collide between GHC and GHE, So we map them to AthenianAppIDs instead.
type AthenianAppID int64

// AppID is a Github App ID.
type AppID int64

// AppContext contains metadata that helps to identify a specific Github application.
type AppContext struct {
	// AthenianAppID is an Athenian App ID for the Github App responsible for the event.
	AthenianAppID AthenianAppID `json:"athenian_app_id,omitempty"`
	// AppID is a Github App ID. It only makes sense in the combination with AthenianAppID.
	AppID AppID `json:"app_id,omitempty"`
}

func (app AppContext) Validate() error {
	if app.AppID <= 0 {
		return ErrNoAppMeta
	} else if app.AthenianAppID <= 0 {
		return ErrNoAppMeta
	}
	return nil
}

type Application struct {
	AppContext
	AppSlug string
	Secret  string
}

// InstallID is a Github App installation ID.
type InstallID int64

// InstallContext contains metadata that helps to identify a specific Github installation.
type InstallContext struct {
	AppContext
	AccountID AccID     `json:"athenian_acc_id,omitempty"`
	InstallID InstallID `json:"install_id,omitempty"`
}

type EventID string

// WithEvent sets a Github event ID for the current context.
func WithEvent(ctx context.Context, id EventID) context.Context {
	return report.WithStringValue(ctx, "webhook.event_id", string(id))
}

// EventContext contains metadata that helps to identify a Github event for a specific installation.
type EventContext struct {
	InstallContext
	EventID EventID `json:"event_id,omitempty"`
}

const (
	FeatureGHE = "athenian.github.ghe"
)

// Feature is a feature flag in the format "example.feature_one".
type Feature string

// Features is a set of Github feature flags.
type Features []Feature

// IsSet checks if a feature flag is set.
func (f Features) IsSet(v Feature) bool {
	for _, f := range f {
		if f == v {
			return true
		}
	}
	return false
}

// Strings converts feature flags to strings.
func (f Features) Strings() []string {
	vals := make([]string, 0, len(f))
	for _, v := range f {
		vals = append(vals, string(v))
	}
	return vals
}

// Merge creates a new feature set as a combination of a given two.
func (f Features) Merge(f2 Features) Features {
	var f3 Features
	seen := make(map[Feature]struct{})
	for _, v := range f {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		f3 = append(f3, v)
	}
	for _, v := range f2 {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		f3 = append(f3, v)
	}
	return f3
}

type keyFeatures struct{}

// WithFeatures sets a Github feature flags for the current context.
func WithFeatures(ctx context.Context, list ...Feature) context.Context {
	nlist := GetFeatures(ctx).Merge(Features(list))
	ctx = context.WithValue(ctx, keyFeatures{}, nlist)
	return report.WithStringValues(ctx, "github.features", nlist.Strings())
}

// GetFeatures lists Github feature flags is set on the context.
func GetFeatures(ctx context.Context) Features {
	list, _ := ctx.Value(keyFeatures{}).(Features)
	return list
}

// FeatureIsSet check if Github feature flag is set on the context.
func FeatureIsSet(ctx context.Context, f Feature) bool {
	return GetFeatures(ctx).IsSet(f)
}

// NodeID is a Github's GraphQL node ID.
type NodeID string

func NodeIDsToStrings(ids []NodeID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

type accIDKey struct{}
type appIDKey struct{}
type installIDKey struct{}

func (id AccID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id InstallID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// WithApplication sets a Github App for the current context.
func WithApplication(ctx context.Context, actx AppContext) context.Context {
	if actx.AthenianAppID != 0 {
		ctx = report.WithInt64Value(ctx, "athenian.app_id", int64(actx.AthenianAppID))
	}
	if actx.AppID != 0 {
		ctx = report.WithInt64Value(ctx, "github.app_id", int64(actx.AppID))
	}
	ctx = context.WithValue(ctx, appIDKey{}, actx)
	return ctx
}

// ApplicationContext returns a Github App context, if any.
func ApplicationContext(ctx context.Context) (AppContext, bool) {
	actx, ok := ctx.Value(appIDKey{}).(AppContext)
	return actx, ok
}

// WithAccount sets an Athenian Github Account ID for the current context.
func WithAccount(ctx context.Context, id AccID) context.Context {
	ctx = report.WithInt64Value(ctx, "athenian.github_acc_id", int64(id))
	ctx = context.WithValue(ctx, accIDKey{}, id)
	return ctx
}

// AccountID returns an Athenian Github Account ID context, if any.
func AccountID(ctx context.Context) (AccID, bool) {
	actx, ok := ctx.Value(accIDKey{}).(AccID)
	return actx, ok
}

// AccountIDStr returns an Athenian Github Account ID context as a string.
func AccountIDStr(ctx context.Context) string {
	if acc, ok := AccountID(ctx); ok {
		return acc.String()
	}
	return ""
}

// WithInstallation sets a Github App installation for the current context.
func WithInstallation(ctx context.Context, ictx InstallContext) context.Context {
	ctx = WithApplication(ctx, ictx.AppContext)
	if ictx.InstallID != 0 {
		ctx = report.WithInt64Value(ctx, "github.install_id", int64(ictx.InstallID))
	}
	ctx = context.WithValue(ctx, installIDKey{}, ictx)
	if ictx.AccountID != 0 {
		ctx = WithAccount(ctx, ictx.AccountID)
	}
	return ctx
}

// InstallationContext returns a Github App installation context, if any.
func InstallationContext(ctx context.Context) (InstallContext, bool) {
	ictx, ok := ctx.Value(installIDKey{}).(InstallContext)
	return ictx, ok
}

// InstallationIDStr returns a Github App installation ID from the context as a string.
func InstallationIDStr(ctx context.Context) string {
	if ictx, ok := InstallationContext(ctx); ok {
		return ictx.InstallID.String()
	}
	return ""
}

// WithNodeID sets a Github node ID for the current context.
func WithNodeID(ctx context.Context, id string) context.Context {
	return report.WithStringValue(ctx, "github.node_id", id)
}

// WithNodeFields sets a Github node fields for the current context.
func WithNodeFields(ctx context.Context, fields []string) context.Context {
	return report.WithStringValues(ctx, "github.node_fields", fields)
}

// WithParentNode sets a parent Github node ID for the current context.
func WithParentNode(ctx context.Context, id, field string) context.Context {
	ctx = report.WithStringValue(ctx, "github.parent_node_id", id)
	if field != "" {
		ctx = report.WithStringValue(ctx, "github.parent_node_field", field)
	}
	return ctx
}

// WithNodeType sets a Github node type for the current context.
func WithNodeType(ctx context.Context, typ string) context.Context {
	return report.WithStringValue(ctx, "github.node_type", typ)
}

// WithNodeIDs sets a set of Github node IDs for the current context.
func WithNodeIDs(ctx context.Context, ids []string) context.Context {
	ctx = report.WithStringValues(ctx, "github.node_ids", ids)
	ctx = report.WithIntValue(ctx, "github.node_batch", len(ids))
	return ctx
}

type AppLister interface {
	ListApplications(ctx context.Context) ([]Application, error)
}
