package handlera

import (
	"context"
	"emobile/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

type InterStruct struct {
	Inter SubscriptionStorage
	//DB    *pgxpool.Pool
	//	DB *pgx.Conn
}

type SubscriptionStorage interface {
	AddSub(ctx context.Context, sub models.Subscription) (cTag pgconn.CommandTag, err error)
	ListSub(ctx context.Context, pageSize, offset int) (subs []models.Subscription, err error)
	ReadSub(ctx context.Context, sub models.Subscription) (subs []models.Subscription, err error)
	UpdateSub(ctx context.Context, sub models.Subscription) (cTag pgconn.CommandTag, err error)
	DeleteSub(ctx context.Context, sub models.Subscription) (cTag pgconn.CommandTag, err error)
	SumSub(ctx context.Context, sub models.Subscription) (summa int64, err error)
	Close()
}
 
