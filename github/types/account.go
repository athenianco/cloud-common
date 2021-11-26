package types

import (
	"context"
	"net/url"
	"strings"
)

// AccountNameFromURL returns a short name of the account based on the Github App installation URL.
// It properly accounts for personal installations as well as Github Enterprise.
func AccountNameFromURL(addr string, owner string) string {
	if addr == "" {
		return owner
	}
	u, err := url.Parse(addr)
	if err != nil {
		return ""
	}
	const github = "github.com"
	sub := strings.SplitN(u.Path, "/", 6)
	if len(sub) < 3 {
		// handle partial URLs (just in case)
		if u.Host != github {
			return strings.TrimSuffix(u.Host+"/"+owner, "/")
		}
		return owner
	}
	if sub[1] != "organizations" {
		// https://github.com/settings/installations/<id>
		return owner
	}
	// https://<domain>/organizations/<org>/settings/installations/<id>
	if u.Host != github {
		return u.Host + "/" + sub[2]
	}
	return sub[2]
}

// Account is an Athenian Github installation account.
type Account struct {
	InstallContext
	FetchWith InstallID
	URL       string
	Name      string
	Endpoint  string
	Active    bool
	Suspended bool
	Features  Features
}

// GetName returns a short account name. It will compute it from URL if Name is not set.
func (inst *Account) GetName() string {
	if inst.Name != "" {
		return inst.Name
	}
	return AccountNameFromURL(inst.URL, "")
}

// AccountGetter is a minimal interface for getting Account records.
type AccountGetter interface {
	// GetAccountById returns account record based on the installation context.
	// Implementations should consider at least AccID and AthenianAppID + InstallID keys.
	// It returns dbs.ErrNotFound is record doesn't exist.
	GetAccountById(ctx context.Context, ictx InstallContext) (*Account, error)
}
