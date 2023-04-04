package types

import (
	"context"
	"time"

	gtypes "github.com/athenianco/cloud-common/github/types"
	jtypes "github.com/athenianco/cloud-common/jira/types"
)

type AccountID int64
type AccountFeature string
type GithubAccountID = gtypes.AccID
type JiraAccountID = jtypes.AccID

const ApiChannelFeature = "api_channel"

type Account struct {
	ID         AccountID
	CreatedAt  time.Time
	Secret     string
	SecretSalt int
	ExpiresAt  time.Time
}

type InstaflowStatus struct {
	AccID                AccountID
	AccountCreated       time.Time
	FetchStarted         time.Time
	FetchCompleted       time.Time
	ConsistencyStarted   time.Time
	ConsistencyCompleted time.Time
	PrecomputeStarted    time.Time
	PrecomputeCompleted  time.Time
	Status               string
}

type Database interface {
	GetAccount(ctx context.Context, id AccountID) (*Account, error)
	GetAccountBySecret(ctx context.Context, secret string) (*Account, error)
	ListAccounts(ctx context.Context) ([]Account, error)
	// DEV-3198
	SetAccountFeature(ctx context.Context, id AccountID, feature AccountFeature, parameters interface{}) error
	UnsetAccountFeature(ctx context.Context, id AccountID, feature AccountFeature) error

	CreateJiraToAthenian(ctx context.Context, jid JiraAccountID, aid AccountID) error
	DeleteJiraToAthenian(ctx context.Context, aid AccountID) error
	JiraToAthenian(ctx context.Context, id JiraAccountID) (AccountID, error)
	AthenianToJira(ctx context.Context, id AccountID) ([]JiraAccountID, error)

	GithubToAthenian(ctx context.Context, id GithubAccountID) (AccountID, error)
	AthenianToGithub(ctx context.Context, id AccountID) ([]GithubAccountID, error)

	GetInstaflowStatus(ctx context.Context, accID AccountID) (*InstaflowStatus, error)
	UpdateInstaflowStatus(ctx context.Context, accID AccountID, timestamp time.Time, status string) error

	Close() error
}
