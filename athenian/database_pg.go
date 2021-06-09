package athenian

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/athenianco/cloud-common/dbs"
)

const (
	pgConnMaxLifetime = time.Minute
	pgConnMaxIdleTime = 30 * time.Second
)

// OpenDatabaseFromEnv opens default postgres database based on environment variable:
// STATE_DATABASE_URI
func OpenDatabaseFromEnv() (Database, error) {
	const dbEnv = "STATE_DATABASE_URI"
	dbURI := os.Getenv(dbEnv)
	if dbURI == "" {
		return nil, errors.New(dbEnv + " is not set")
	}
	return Open(context.Background(), processAddress(dbURI))
}

// Open creates a state database based on Postgres.
func Open(ctx context.Context, addr string) (Database, error) {
	config, err := pgxpool.ParseConfig(addr)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.PreferSimpleProtocol = true
	config.MaxConnLifetime = pgConnMaxLifetime
	config.MaxConnIdleTime = pgConnMaxIdleTime

	conn, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return &database{db: conn}, nil
}

type database struct {
	db *pgxpool.Pool
}

func processAddress(addr string) string {
	return strings.Replace(addr, "&binary_parameters=yes", "", -1)
}

func scanAccount(sc dbs.Scanner) (Account, error) {
	var (
		id      int64
		created time.Time
		secret  string
		salt    int
	)
	err := sc.Scan(&id, &created, &secret, &salt)
	return Account{
		ID:         AccountID(id),
		CreatedAt:  created,
		Secret:     secret,
		SecretSalt: salt,
	}, err
}

const accountColumns = `id, created_at, secret, secret_salt`

func (db *database) GetAccount(ctx context.Context, id AccountID) (*Account, error) {
	row := db.db.QueryRow(ctx, `SELECT `+accountColumns+` FROM public.accounts WHERE id = $1`, int64(id))
	acc, err := scanAccount(row)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (db *database) GetAccountBySecret(ctx context.Context, secret string) (*Account, error) {
	row := db.db.QueryRow(ctx, `SELECT `+accountColumns+` FROM public.accounts WHERE secret = $1`, secret)
	acc, err := scanAccount(row)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (db *database) ListAccounts(ctx context.Context) ([]Account, error) {
	rows, err := db.db.Query(ctx, `SELECT `+accountColumns+` FROM public.accounts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Account
	for rows.Next() {
		acc, err := scanAccount(rows)
		if err != nil {
			return out, err
		}
		out = append(out, acc)
	}
	return out, rows.Err()
}

func (db *database) CreateJiraToAthenian(ctx context.Context, jid JiraAccountID, aid AccountID) error {
	_, err := db.db.Exec(ctx, `INSERT INTO public.account_jira_installations(id, account_id)
		VALUES($1, $2)`, int64(jid), int64(aid))
	return err
}

func (db *database) DeleteJiraToAthenian(ctx context.Context, aid AccountID) error {
	_, err := db.db.Exec(ctx, `DELETE FROM public.account_jira_installations where account_id = $1;`, int64(aid))
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	return err
}

func (db *database) JiraToAthenian(ctx context.Context, id JiraAccountID) (AccountID, error) {
	var accID int64
	err := db.db.QueryRow(ctx, `SELECT account_id FROM public.account_jira_installations WHERE id = $1`, int64(id)).Scan(&accID)
	if err == pgx.ErrNoRows {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}
	return AccountID(accID), nil
}

func (db *database) AthenianToJira(ctx context.Context, id AccountID) ([]JiraAccountID, error) {
	rows, err := db.db.Query(ctx, `SELECT id FROM public.account_jira_installations WHERE account_id = $1`, int64(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []JiraAccountID
	for rows.Next() {
		var rid int64
		if err := rows.Scan(&rid); err != nil {
			return out, err
		}
		out = append(out, JiraAccountID(rid))
	}
	return out, rows.Err()
}

func (db *database) GithubToAthenian(ctx context.Context, id GithubAccountID) (AccountID, error) {
	var accID int64
	err := db.db.QueryRow(ctx, `SELECT account_id FROM public.account_github_accounts WHERE id = $1`, int64(id)).Scan(&accID)
	if err == pgx.ErrNoRows {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}
	return AccountID(accID), nil
}

func (db *database) AthenianToGithub(ctx context.Context, id AccountID) ([]GithubAccountID, error) {
	rows, err := db.db.Query(ctx, `SELECT id FROM public.account_github_accounts WHERE account_id = $1`, int64(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GithubAccountID
	for rows.Next() {
		var rid int64
		if err := rows.Scan(&rid); err != nil {
			return out, err
		}
		out = append(out, GithubAccountID(rid))
	}
	return out, rows.Err()
}

func (db *database) Close() error {
	db.db.Close()
	return nil
}
