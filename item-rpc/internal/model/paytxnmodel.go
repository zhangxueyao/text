package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PayTxnModel = (*customPayTxnModel)(nil)

type (
	// PayTxnModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPayTxnModel.
	PayTxnModel interface {
		payTxnModel
		withSession(session sqlx.Session) PayTxnModel
	}

	customPayTxnModel struct {
		*defaultPayTxnModel
	}
)

// NewPayTxnModel returns a model for the database table.
func NewPayTxnModel(conn sqlx.SqlConn) PayTxnModel {
	return &customPayTxnModel{
		defaultPayTxnModel: newPayTxnModel(conn),
	}
}

func (m *customPayTxnModel) withSession(session sqlx.Session) PayTxnModel {
	return NewPayTxnModel(sqlx.NewSqlConnFromSession(session))
}
