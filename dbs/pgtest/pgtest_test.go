package pgtest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLSplit(t *testing.T) {
	stmts := sqlSplitStmts([]byte(`
CREATE OR REPLACE FUNCTION public.github_repo_org_name_from_full_name(full_name text) RETURNS text AS $$
BEGIN
    RETURN split_part(full_name, '/', 1);
END; $$ LANGUAGE plpgsql;

DROP TABLE IF EXISTS public.github_repositories;
DROP TABLE IF EXISTS public.github_users;
`))
	require.Equal(t, []string{
		`CREATE OR REPLACE FUNCTION public.github_repo_org_name_from_full_name(full_name text) RETURNS text AS $$
BEGIN
    RETURN split_part(full_name, '/', 1);
END; $$ LANGUAGE plpgsql`,
		`DROP TABLE IF EXISTS public.github_repositories`,
		`DROP TABLE IF EXISTS public.github_users`,
	}, stmts)
}
