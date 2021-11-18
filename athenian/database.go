package athenian

import (
	"github.com/athenianco/cloud-common/athenian/types"
	"github.com/athenianco/cloud-common/dbs"
	gtypes "github.com/athenianco/cloud-common/github/types"
	jtypes "github.com/athenianco/cloud-common/jira/types"
)

type AccountID = types.AccountID
type AccountFeature = types.AccountFeature
type GithubAccountID = gtypes.AccID
type JiraAccountID = jtypes.AccID

var ErrNotFound = dbs.ErrNotFound

type Account = types.Account

type Database = types.Database
