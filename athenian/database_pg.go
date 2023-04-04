package athenian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		expires time.Time
	)
	err := sc.Scan(&id, &created, &secret, &salt, &expires)
	return Account{
		ID:         AccountID(id),
		CreatedAt:  created,
		Secret:     secret,
		SecretSalt: salt,
		ExpiresAt:  expires,
	}, err
}

const accountColumns = `id, created_at, secret, secret_salt, expires_at`

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

func (db *database) getAccountFeatureID(ctx context.Context, feature AccountFeature) (int64, error) {
	row := db.db.QueryRow(ctx, `SELECT id FROM features WHERE name = $1`, feature)
	var id int64
	err := row.Scan(&id)
	if err == pgx.ErrNoRows {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *database) SetAccountFeature(ctx context.Context, id AccountID, feature AccountFeature, parameters interface{}) error {
	paramsEncData, err := json.Marshal(parameters)
	if err != nil {
		return err
	}

	fid, err := db.getAccountFeatureID(ctx, feature)
	if err != nil {
		return err
	}
	_, err = db.db.Exec(ctx, `INSERT 
INTO account_features (account_id, feature_id, enabled, parameters, updated_at) 
VALUES($1, $2, true, $3, NOW())
ON CONFLICT(account_id, feature_id)
DO UPDATE SET (account_id, feature_id, enabled, parameters, updated_at) = ($1, $2, true, $3, NOW())`, id, fid, string(paramsEncData))
	return err
}

func (db *database) UnsetAccountFeature(ctx context.Context, id AccountID, feature AccountFeature) error {
	fid, err := db.getAccountFeatureID(ctx, feature)
	if err != nil {
		return err
	}
	_, err = db.db.Exec(ctx, `DELETE FROM account_features WHERE account_id = $1 AND feature_id = $2`, id, fid)
	return err
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

func (db *database) GetInstaflowStatus(ctx context.Context, accID AccountID) (*InstaflowStatus, error) {
	return db.getInstaflowStatus(ctx, accID)
}

func (db *database) getInstaflowStatus(ctx context.Context, accID AccountID) (*InstaflowStatus, error) {
	row := db.db.QueryRow(
		ctx,
		`SELECT account_id, account_created, fetch_started, fetch_completed, consistency_started, consistency_completed, precompute_started, precompute_completed, current_status
		FROM public.installation_progress WHERE account_id = $1`,
		int64(accID),
	)
	instaflowStatus, err := scanInstaflowStatus(row)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}
	return &instaflowStatus, err
}

func scanInstaflowStatus(sc dbs.Scanner) (InstaflowStatus, error) {
	var (
		accID                int64
		accountCreated       time.Time
		fetchStarted         time.Time
		fetchCompleted       time.Time
		consistencyStarted   time.Time
		consistencyCompleted time.Time
		precomputeStarted    time.Time
		precomputeCompleted  time.Time
		status               string
	)
	err := sc.Scan(&accID, &accountCreated, &fetchStarted, &fetchCompleted, &consistencyStarted, &consistencyCompleted, &precomputeStarted, &precomputeCompleted, &status)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}

	return InstaflowStatus{
		AccID:                AccountID(accID),
		AccountCreated:       accountCreated,
		FetchStarted:         fetchStarted,
		FetchCompleted:       fetchCompleted,
		ConsistencyStarted:   consistencyStarted,
		ConsistencyCompleted: consistencyCompleted,
		PrecomputeStarted:    precomputeStarted,
		PrecomputeCompleted:  precomputeCompleted,
		Status:               status,
	}, err
}

func (db *database) UpdateInstaflowStatus(ctx context.Context, accID AccountID, timestamp time.Time, status string) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var qu string

	ext, err := db.getInstaflowStatus(ctx, accID)
	if err == ErrNotFound {
		qu = fmt.Sprintf(`INSERT INTO public.installation_progress (account_id, %s, current_status)
		VALUES ($1, $2, $3)
		ON CONFLICT (account_id) DO UPDATE SET %s = $2, current_status = $3 where account_id = $1`, status, status)
	} else if ext.Status != status {
		qu = fmt.Sprintf("UPDATE public.installation_progress SET %s = $2, current_status = $3 where account_id = $1", status)
	} else {
		// up to date
		return nil
	}

	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	_, err = tx.Exec(ctx, qu, int64(accID), timestamp, status)
	if err != nil {
		return err
	}
	return nil
}

func (db *database) Close() error {
	db.db.Close()
	return nil
}
