package model

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OutboxEvent struct {
	EventID     int64           `db:"event_id"`
	Aggregate   string          `db:"aggregate"`
	AggregateID int64           `db:"aggregate_id"`
	EventType   string          `db:"event_type"`
	Payload     json.RawMessage `db:"payload"`
	Status      int             `db:"status"` // 0=NEW,1=PUBLISHED
	CreatedAt   time.Time       `db:"created_at"`
}

type OutboxModel interface {
	TxInsert(ctx context.Context, s sqlx.Session, evt *OutboxEvent) error
	ListNew(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkPublished(ctx context.Context, id int64) error
}

type defaultOutboxModel struct{ conn sqlx.SqlConn }

func NewOutboxModel(conn sqlx.SqlConn) OutboxModel { return &defaultOutboxModel{conn: conn} }

func (m *defaultOutboxModel) TxInsert(ctx context.Context, s sqlx.Session, e *OutboxEvent) error {
	_, err := s.ExecCtx(ctx, `insert into outbox_events(event_id,aggregate,aggregate_id,event_type,payload,status)
values(?,?,?,?,?,0)`, e.EventID, e.Aggregate, e.AggregateID, e.EventType, e.Payload)
	return err
}

func (m *defaultOutboxModel) ListNew(ctx context.Context, limit int) ([]OutboxEvent, error) {
	var rows []OutboxEvent
	err := m.conn.QueryRowsCtx(ctx, &rows, `select event_id,aggregate,aggregate_id,event_type,payload,status,created_at
from outbox_events where status=0 order by created_at asc limit ?`, limit)
	return rows, err
}

func (m *defaultOutboxModel) MarkPublished(ctx context.Context, id int64) error {
	_, err := m.conn.ExecCtx(ctx, `update outbox_events set status=1 where event_id=?`, id)
	return err
}
