package dbase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"emobile/internal/models"
)

// Структура для базы данных.
type DBstruct struct {
	DB *pgxpool.Pool
	//	DB *pgx.Conn
}

func NewPostgresPool(ctx context.Context, DSN string) (*DBstruct, error) {

	poolConfig, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	return &DBstruct{DB: pool}, nil
}

func Ping(ctx context.Context) error {
	dataBase, err := NewPostgresPool(ctx, models.DSN)
	if err != nil {
		return err
	}
	defer dataBase.DB.Close()

	err = dataBase.DB.Ping(ctx) // база то открыта ...
	if err != nil {
		models.Logger.Error("No PING ", "error", err.Error())
		return fmt.Errorf("no ping %w", err)
	}
	return nil
}

// AddSub добавление подписки в Базу Данных.
func (dataBase *DBstruct) AddSub(ctx context.Context, sub models.Subscription) (cTag pgconn.CommandTag, err error) {

	// валидность полей подписки проверена в CreateHandler

	order := "INSERT INTO subscriptions(service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5) ;"
	cTag, err = dataBase.DB.Exec(ctx, order, sub.Service_name, sub.Price, sub.User_id, sub.Sdt, sub.Edt)

	return
}

func (dataBase *DBstruct) ListSub(ctx context.Context, pageSize, offset int) (subs []models.Subscription, err error) {

	order := "SELECT service_name, price, user_id, start_date, end_date FROM subscriptions LIMIT $1 OFFSET $2"
	rows, err := dataBase.DB.Query(ctx, order, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// в таблице нет NULL значений, поэтому сканируем напрямую

	for rows.Next() {
		sub := models.Subscription{}
		if err := rows.Scan(&sub.Service_name, &sub.Price, &sub.User_id, &sub.Sdt, &sub.Edt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return
}

func (dataBase *DBstruct) ReadSub(ctx context.Context, sub models.Subscription) (subs []models.Subscription, err error) {

	order := `
		SELECT service_name, price, user_id, start_date, end_date FROM subscriptions WHERE 
		(service_name=$1 OR service_name='')  
		AND ($2::int = 0 OR price = $2::int)
		AND (user_id = NULLIF($3, '')::uuid OR $3 = '')
		AND ( start_date <= $4 OR $4 = '0001-01-01 00:00:00' ) 
		AND ( end_date >= $5 ) ;
	`

	rows, err := dataBase.DB.Query(ctx, order, sub.Service_name, sub.Price, sub.User_id, sub.Sdt, sub.Edt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		sub := models.Subscription{}
		var sdt, edt sql.NullTime
		if err := rows.Scan(&sub.Service_name, &sub.Price, &sub.User_id, &sdt, &edt); err != nil {
			return nil, err
		}
		sub.Sdt = sdt.Time
		sub.Edt = edt.Time
		subs = append(subs, sub)
	}

	return
}

// UpdateSub - обновление данных подписки
func (dataBase *DBstruct) UpdateSub(ctx context.Context, sub models.Subscription) (cTag pgconn.CommandTag, err error) {

	order := `
		UPDATE subscriptions SET 
		-- если sub.Price == 0 оставить в таблице прежнее значение price
		price = COALESCE(NULLIF($1, 0), price),
		start_date=$2, 
		end_date=$3 
		WHERE service_name=$4 AND user_id=$5::uuid;
	`

	cTag, err = dataBase.DB.Exec(ctx, order, sub.Price, sub.Sdt, sub.Edt, sub.Service_name, sub.User_id)

	return
}

func (dataBase *DBstruct) DeleteSub(ctx context.Context, sub models.Subscription) (cTag pgconn.CommandTag, err error) {

	order := `
		DELETE FROM subscriptions
		WHERE
		(service_name = $1 OR $1 = '')
		AND (price = $2 OR $2::int = 0)
		AND (user_id = NULLIF($3, '')::uuid OR $3 = '')
		AND (start_date <= $4 OR $4 = '0001-01-01 00:00:00')
		AND (end_date >= $5 OR $5 = '0001-01-01 00:00:00' );
	`

	cTag, err = dataBase.DB.Exec(ctx, order, sub.Service_name, sub.Price, sub.User_id, sub.Sdt, sub.Edt)
	if err != nil {
		models.Logger.Error("Delete", "", err.Error())
	}

	return
}

func (dataBase *DBstruct) SumSub(ctx context.Context, sub models.Subscription) (summa int64, err error) {

	//  если конечная дата подписки не задана - устанавлiваем в максимально возможное значение
	//  			Жаль только — жить в эту пору прекрасную уж не придется — ни мне, ни тебе ©
	if sub.Edt.IsZero() {
		sub.Edt = time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	}

	order := `
		WITH filtered_subscriptions AS (
			SELECT id, service_name, price, start_date, end_date
			FROM subscriptions
			WHERE
				($1 = '' OR service_name = $1) 
				AND (user_id = NULLIF($2, '')::uuid OR $2 = '')
				-- AND ($2 = '' OR user_id = $2::UUID)
				AND GREATEST($3::DATE, start_date) <= LEAST($4::DATE, end_date)
		)
		SELECT SUM(price * (
			EXTRACT(YEAR FROM AGE(LEAST($4::DATE, end_date), GREATEST($3::DATE, start_date))) * 12 +
			EXTRACT(MONTH FROM AGE(LEAST($4::DATE, end_date), GREATEST($3::DATE, start_date))) + 1
		))
		FROM filtered_subscriptions;
	`

	var nullsum sql.NullInt64

	row := dataBase.DB.QueryRow(ctx, order, sub.Service_name, sub.User_id, sub.Sdt, sub.Edt)

	err = row.Scan(&nullsum)
	if err != nil {
		return 0, err
	}
	if nullsum.Valid {
		summa = nullsum.Int64
		return
	}

	return 0, sql.ErrNoRows
}

func (dataBase *DBstruct) Close() {
	dataBase.DB.Close()
}
 