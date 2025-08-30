package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Item struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
	// ... 其他字段
}

type ItemModel interface {
	FindOne(ctx context.Context, id int64) (*Item, error)
	TxUpdate(ctx context.Context, s sqlx.Session, it *Item) (sql.Result, error)
}

type defaultItemModel struct{ conn sqlx.SqlConn }

func NewItemModel(conn sqlx.SqlConn) ItemModel { return &defaultItemModel{conn: conn} }

func (m *defaultItemModel) FindOne(ctx context.Context, id int64) (*Item, error) {
	var it Item
	err := m.conn.QueryRowCtx(ctx, &it, "select id,name from item-api where id=?", id)
	return &it, err
}

func (m *defaultItemModel) TxUpdate(ctx context.Context, s sqlx.Session, it *Item) (sql.Result, error) {
	return s.ExecCtx(ctx, "update item-api set name=? where id=?", it.Name, it.Id)
}
