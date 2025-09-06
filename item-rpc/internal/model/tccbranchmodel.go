package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TccBranchModel = (*customTccBranchModel)(nil)

type (
	// TccBranchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTccBranchModel.
	TccBranchModel interface {
		tccBranchModel
		withSession(session sqlx.Session) TccBranchModel
	}

	customTccBranchModel struct {
		*defaultTccBranchModel
	}
)

// NewTccBranchModel returns a model for the database table.
func NewTccBranchModel(conn sqlx.SqlConn) TccBranchModel {
	return &customTccBranchModel{
		defaultTccBranchModel: newTccBranchModel(conn),
	}
}

func (m *customTccBranchModel) withSession(session sqlx.Session) TccBranchModel {
	return NewTccBranchModel(sqlx.NewSqlConnFromSession(session))
}
