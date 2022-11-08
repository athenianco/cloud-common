package types

import (
	"context"
)

type ShardID int

type Shard struct {
	ID        ShardID
	AccID     AccID
	AppID     AthenianAppID
	InstallID InstallID
}

type ShardsDatabase interface {
	CreateShard(ctx context.Context, shard Shard) error
	GetShardID(ctx context.Context, appID AthenianAppID, installID InstallID) (ShardID, error)
	GetShardIDByAcc(ctx context.Context, accID AccID) (ShardID, error)
	ListShards(ctx context.Context) ([]Shard, error)
}
