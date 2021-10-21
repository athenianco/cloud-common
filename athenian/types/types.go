package types

import (
	"context"
	"time"

	gtypes "github.com/athenianco/cloud-common/github/types"
	jtypes "github.com/athenianco/cloud-common/jira/types"
)

type AccountID int64
type GithubAccountID = gtypes.AccID
type JiraAccountID = jtypes.AccID

type Account struct {
	ID         AccountID
	CreatedAt  time.Time
	Secret     string
	SecretSalt int
	ExpiresAt  time.Time
}

type Database interface {
	GetAccount(ctx context.Context, id AccountID) (*Account, error)
	GetAccountBySecret(ctx context.Context, secret string) (*Account, error)
	ListAccounts(ctx context.Context) ([]Account, error)

	CreateJiraToAthenian(ctx context.Context, jid JiraAccountID, aid AccountID) error
	DeleteJiraToAthenian(ctx context.Context, aid AccountID) error
	JiraToAthenian(ctx context.Context, id JiraAccountID) (AccountID, error)
	AthenianToJira(ctx context.Context, id AccountID) ([]JiraAccountID, error)

	GithubToAthenian(ctx context.Context, id GithubAccountID) (AccountID, error)
	AthenianToGithub(ctx context.Context, id AccountID) ([]GithubAccountID, error)

	Close() error
}
