package database

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/BelyaevEI/gophermart/internal/models"
)

type Database struct {
	database *sql.DB
}

func NewConnect(DBpath string) (*Database, error) {

	db, err := sql.Open("pgx", DBpath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS registation" +
		"(login text NOT NULL, password text NOT NULL, secretkey text NOT NULL, token text NOT NULL)")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS orders" +
		"(token text NOT NULL, num_orders int NOT NULL, num_score money NOT NULL DEFAULT 0, status text NOT NULL DEFAULT NEW, upload timestamp WITH TIME ZONE NOT NULL)")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS scores" +
		"(token text NOT NULL, scores money NOT NULL)")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS withdrawn" +
		"token text NOT NULL, withdrawn money NOT NULL, num_order int NOT NULL processed timestamp WITH TIME ZONE")
	if err != nil {
		return nil, err
	}

	return &Database{
		database: db,
	}, nil
}

func (d *Database) SaveUser(ctx context.Context, login, password, key string) error {
	_, err := d.database.ExecContext(ctx, "INSERT INTO registration(login, password, secretkey)"+
		"values($1, $2, $3)", login, password, key)
	return err
}

func (d *Database) SaveToken(ctx context.Context, token, key string) error {
	_, err := d.database.ExecContext(ctx, "INSERT INTO auth(token, secretkey) values($1, $2)", token, key)
	return err
}

func (d *Database) CheckUserLogin(ctx context.Context, login string) error {
	_, err := d.database.ExecContext(ctx, "SELECT login FROM registration WHERE login = $1", login)
	return err
}

func (d *Database) GetUser(ctx context.Context, login string) *sql.Row {
	return d.database.QueryRowContext(ctx, "SELECT login, password, secretkey, token"+
		"FROM registration WHERE login = $1", login)
}

func (d *Database) GetKey(ctx context.Context, token string) *sql.Row {
	return d.database.QueryRowContext(ctx, "SELECT secretkey FROM registration WHERE token = $1", token)
}

func (d *Database) CheckOrder(ctx context.Context, numOrder int, token string) int {
	number := 0
	row := d.database.QueryRowContext(ctx, "SELECT num_order FROM orders WHERE token = $1", token)
	if err := row.Scan(&number); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return http.StatusOK
		}
		return 0
	}

	row = d.database.QueryRowContext(ctx, "SELECT num_order FROM orders WHERE num_order = $1", numOrder)
	if err := row.Scan(&number); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return http.StatusConflict
		}
		return 0
	}
	return 1
}

func (d *Database) Upload2System(ctx context.Context, numOrder int, token string) error {
	_, err := d.database.ExecContext(ctx, "INSERT INTO orders(token, num_order, upload) values($1, $2, $3)", token, numOrder, time.Now())
	return err
}

func (d *Database) GetNewOrders() ([]models.Orders, error) {
	orders := make([]models.Orders, 0)

	rows, err := d.database.Query("SELECT token, num_orders FROM orders WHERE status = NEW")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var order models.Orders

		err = rows.Scan(&order.Token, &order.NumOrder)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (d *Database) UpdateOrder(order models.Order, token string) error {
	_, err := d.database.Exec("UPDATE orders SET status = $1 WHERE token = $2 AND num_orders = $3", order.Status, order.Order, token)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateUserScores(point int, token string) error {

	scores := 0
	row, err := d.database.Query("SELECT scores FROM scores WHERE token = $1", token)
	if err != nil {
		if err := row.Scan(&scores); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}
	}

	scores += point

	_, err = d.database.Exec("UPDATE scores SET scores = $1 WHERE token = $2", scores, token)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetOrdersUser(ctx context.Context, token string) ([]models.UserOrders, error) {
	orders := make([]models.UserOrders, 0)

	rows, err := d.database.QueryContext(ctx, "SELECT num_orders, num_score, status, upload FROM orders WHERE token = $1", token)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			order models.UserOrders
			time  time.Time
		)

		err = rows.Scan(&order.NumOrders, &order.Accrual, &order.Status, &time)
		if err != nil {
			return nil, err
		}

		// Convert time to RFC3339
		order.Upload = time.Format("2006-01-02T15:04:05Z07:00")
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (d *Database) GetAccrualUser(ctx context.Context, token string) (float64, error) {
	var scores float64
	row := d.database.QueryRowContext(ctx, "SELECT scores FROM scores WHERE token = $1", token)
	if err := row.Scan(&scores); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}
	return scores, nil
}

func (d *Database) GetWithdrawnUser(ctx context.Context, token string) (float64, error) {
	var withdrawn float64
	row := d.database.QueryRowContext(ctx, "SELECT withdrawn FROM cancellation WHERE token = &1", token)
	if err := row.Scan(&withdrawn); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}
	return withdrawn, nil
}

func (d *Database) UpdateWithdrawn(ctx context.Context, token string, sum float64, order int) error {
	_, err := d.database.Exec("INSERT withdrawn SET withdrawn = $1 processed = $2 num_order = $4 WHERE token = $5", sum, time.Now(), order, token)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetUSerScore(ctx context.Context, token string) (float64, error) {
	var scores float64
	row, err := d.database.Query("SELECT scores FROM scores WHERE token = $1", token)
	if err != nil {
		if err := row.Scan(&scores); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return 0, err
			}
		}
	}
	return scores, nil
}

func (d *Database) UpdateScores(ctx context.Context, token string, score float64) error {
	_, err := d.database.ExecContext(ctx, "UPDATE scores SET scores = $1 WHERE token = $2", score, token)
	return err
}

func (d *Database) GetAllWithdrawn(ctx context.Context, token string) ([]models.AllWithdrawn, error) {
	withdrawns := make([]models.AllWithdrawn, 0)

	rows, err := d.database.QueryContext(ctx, "SELECT withdrawn, num_order, processed FROM withdrawn WHERE token = $1", token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	for rows.Next() {
		var (
			withdrawn models.AllWithdrawn
			time      time.Time
		)

		err = rows.Scan(&withdrawn.Sum, &withdrawn.Order, &time)
		if err != nil {
			return nil, err
		}

		// Convert time to RFC3339
		withdrawn.Processed = time.Format("2006-01-02T15:04:05Z07:00")
		withdrawns = append(withdrawns, withdrawn)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawns, nil
}
