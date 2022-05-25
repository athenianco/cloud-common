package types

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/athenianco/cloud-common/dbs"
	"github.com/athenianco/cloud-common/report"
)

var (
	// ErrNotFound is returned by getter DB functions that return a single object, indicating that an object doesn't exist.
	ErrNotFound = dbs.ErrNotFound
	// ErrNoInstallationMeta is returned if the event doesn't contain required installation-related info.
	ErrNoInstallationMeta = errors.New("no installation metadata")
	// ErrNoAppMeta is returned if the event doesn't contain required app-related info.
	ErrNoAppMeta = errors.New("no application metadata")
	// ErrNoEventMeta is returned if the event doesn't contain required info.
	ErrNoEventMeta = errors.New("no event metadata")
)

// AccID is an Athenian Github Account ID.
// It combines a Github AppID with a Github InstallID and can uniquely identify an installation.
type AccID int64

// AthenianAppID is an Athenian Github App ID.
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

func NewGraphID(id uint64, typ string) GraphID {
	if id == 0 {
		panic("empty ID")
	}
	if typ == "" {
		panic("empty type")
	}
	return GraphID{id: id, typ: typ}
}

func ParseGraphID(s string) (GraphID, error) {
	if s == "" {
		return GraphID{}, nil
	}
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return GraphID{}, errors.New("invalid graph ID format")
	}
	sid, typ := s[:i], s[i+1:]
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		return GraphID{}, err
	}
	return NewGraphID(id, typ), nil
}

type NodeResolver interface {
	// ResolveNode creates or returns existing Athenian node ID for a given type and Github node ID.
	ResolveNode(ctx context.Context, accID AccID, typ, id string) (GraphID, error)
}

// IsSameType checks if all IDs in the slice are of the same node type.
func IsSameType(ids []GraphID) (string, bool) {
	if len(ids) == 0 {
		return "", false
	}
	typ := ids[0].Type()
	for _, id := range ids[1:] {
		if typ != id.Type() {
			return "", false
		}
	}
	return typ, true
}

// WithGraphID sets a graph node ID for the current context.
func WithGraphID(ctx context.Context, id GraphID) context.Context {
	ctx = report.WithStringValue(ctx, "athenian.github.graph_id", id.String())
	ctx = WithNodeType(ctx, id.Type())
	return ctx
}

// WithParentGraphID sets a parent graph node ID for the current context.
func WithParentGraphID(ctx context.Context, id GraphID, field string) context.Context {
	ctx = report.WithStringValue(ctx, "athenian.github.parent_graph_id", id.String())
	if field != "" {
		ctx = report.WithStringValue(ctx, "github.parent_node_field", field)
	}
	return ctx
}

// WithGraphIDs sets graph node IDs for the current context.
func WithGraphIDs(ctx context.Context, ids []GraphID) context.Context {
	if typ, ok := IsSameType(ids); ok {
		ctx = WithNodeType(ctx, typ)
	}
	str := make([]string, 0, len(ids))
	for _, id := range ids {
		str = append(str, id.String())
	}
	ctx = report.WithStringValues(ctx, "athenian.github.graph_id", str)
	return ctx
}

var (
	_ json.Marshaler   = GraphID{}
	_ json.Unmarshaler = (*GraphID)(nil)
)

// GraphID is an integer graph node ID used in Athenian.
type GraphID struct {
	id  uint64
	typ string
}

func (id GraphID) IsZero() bool {
	return id == GraphID{}
}

func (id GraphID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *GraphID) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == nil {
		*id = GraphID{}
		return nil
	}
	v, err := ParseGraphID(*s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

func (id GraphID) ID() uint64 {
	return id.id
}

func (id GraphID) Type() string {
	return id.typ
}

func (id GraphID) Split() (uint64, string) {
	return id.id, id.typ
}

func (id GraphID) String() string {
	return strconv.FormatUint(id.id, 10) + ":" + id.typ
}

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

type AppDatabase interface {
	AppLister
	CreateApplication(ctx context.Context, appID AppID, slug string, secret string) (AppContext, error)
	GetApplication(ctx context.Context, id AthenianAppID) (*Application, error)
}

// AttrsWithAccount sets an Athenian Github Account ID to the attributes map.
func AttrsWithAccount(attrs map[string]string, id AccID) map[string]string {
	if attrs == nil {
		attrs = make(map[string]string)
	}
	if id != 0 {
		attrs["com.athenian.github.acc_id"] = id.String()
	}
	return attrs
}

// AttrsWithApplication sets a Github App for the attributes map.
func AttrsWithApplication(attrs map[string]string, actx AppContext) map[string]string {
	if attrs == nil {
		attrs = make(map[string]string)
	}
	if actx.AthenianAppID != 0 {
		attrs["com.athenian.github.app_id"] = strconv.FormatInt(int64(actx.AthenianAppID), 10)
	}
	if actx.AppID != 0 {
		attrs["com.github.x.app_id"] = strconv.FormatInt(int64(actx.AppID), 10)
	}
	return attrs
}

// AttrsWithInstallation sets a Github App installation for the attributes map.
func AttrsWithInstallation(attrs map[string]string, ictx InstallContext) map[string]string {
	if attrs == nil {
		attrs = make(map[string]string)
	}
	attrs = AttrsWithApplication(attrs, ictx.AppContext)
	attrs = AttrsWithAccount(attrs, ictx.AccountID)
	if ictx.InstallID != 0 {
		attrs["com.github.x.install_id"] = ictx.InstallID.String()
	}
	return attrs
}
