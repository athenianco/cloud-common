package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountName(t *testing.T) {
	for _, c := range []struct {
		URL  string
		Name string
	}{
		{URL: "https://github.com/organizations/myorg/settings/installations/1234567", Name: "myorg"},
		{URL: "https://github.com/organizations/myorg", Name: "myorg"},
		{URL: "https://ghe.athenian.co/organizations/atheniantest/settings/installations/12", Name: "ghe.athenian.co/atheniantest"},
		{URL: "https://ghe.athenian.co/organizations/atheniantest", Name: "ghe.athenian.co/atheniantest"},
		{URL: "https://ghe.athenian.co", Name: "ghe.athenian.co"},
		{URL: "https://github.com/settings/installations/1234567", Name: ""},
	} {
		t.Run(c.Name, func(t *testing.T) {
			inst := Account{URL: c.URL}
			inst.AccountID = 1
			name := inst.GetName()
			require.Equal(t, c.Name, name)
		})
	}
}
