package local

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/athenianco/cloud-common/dbs"
)

type Local struct {
	workDir string
}

func New(workDir string) *Local {
	return &Local{
		workDir: workDir,
	}
}

func (l *Local) GetPrivateKeyData(ctx context.Context, appId int64) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(l.workDir, strconv.Itoa(int(appId))+".pem"))
	if os.IsNotExist(err) {
		return nil, dbs.ErrNotFound
	}
	return data, err
}

func (l *Local) Close() error { return nil }
