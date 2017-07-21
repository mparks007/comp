package circuit

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlLogger struct {
	db *sql.DB
}

func NewMySqlLogger(connectionString string, ctx context.Context) (*MySqlLogger, error) {

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &MySqlLogger{db}, nil
}

func (m *MySqlLogger) Log(cat, data string, ctx context.Context) error {

	if _, err := m.db.ExecContext(ctx, "insert into actionlog (category, detail) values (?, ?)", cat, data); err != nil {
		return err
	}
	return nil
}
