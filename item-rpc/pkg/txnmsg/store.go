package txnmsg

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	stPending = 0
	stSending = 1
	stSent    = 2
	stFailed  = 3
)

type Store struct{ Conn sqlx.SqlConn }

func NewStore(conn sqlx.SqlConn) *Store { return &Store{Conn: conn} }

func (s *Store) AppendTx(ctx context.Context, session sqlx.Session, topic, key string, payload any) error {
	bs, _ := json.Marshal(payload)
	_, err := session.ExecCtx(ctx, `
        INSERT INTO txn_message(topic, msg_key, payload, status, retry_count, available_at, created_at, updated_at)
        VALUES(?, ?, ?, ?, 0, NOW(6), NOW(6), NOW(6))
    `, topic, key, bs, stPending)
	return err
}

func (s *Store) Claim(ctx context.Context, owner string, n int) (int64, error) {
	res, err := s.Conn.ExecCtx(ctx, `
        UPDATE txn_message
        SET status=?, owner=?, updated_at=NOW(6)
        WHERE status=? AND available_at<=NOW(6)
        ORDER BY id LIMIT ?
    `, stSending, owner, stPending, n)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

type Msg struct {
	ID    int64
	Topic string
	Key   string
	Body  []byte
	Retry int
}

func (s *Store) FetchClaimed(ctx context.Context, owner string, n int) ([]Msg, error) {
	var rows []struct {
		ID      int64  `db:"id"`
		Topic   string `db:"topic"`
		MsgKey  string `db:"msg_key"`
		Payload []byte `db:"payload"`
		Retry   int    `db:"retry_count"`
	}
	err := s.Conn.QueryRowsCtx(ctx, &rows, `
        SELECT id, topic, msg_key, payload, retry_count FROM txn_message
        WHERE status=? AND owner=? ORDER BY id LIMIT ?
    `, stSending, owner, n)
	if err != nil {
		return nil, err
	}
	out := make([]Msg, 0, len(rows))
	for _, r := range rows {
		out = append(out, Msg{ID: r.ID, Topic: r.Topic, Key: r.MsgKey, Body: r.Payload, Retry: r.Retry})
	}
	return out, nil
}

func (s *Store) MarkSent(ctx context.Context, id int64) error {
	_, err := s.Conn.ExecCtx(ctx, `UPDATE txn_message SET status=?, updated_at=NOW(6), owner=NULL WHERE id=?`, stSent, id)
	return err
}

func (s *Store) RetryLater(ctx context.Context, id int64, retry int) error {
	backoff := time.Duration(200*(1<<uint(min(retry, 8)))) * time.Millisecond
	_, err := s.Conn.ExecCtx(ctx, `
        UPDATE txn_message SET status=?, retry_count=?, available_at=DATE_ADD(NOW(6), INTERVAL ? MICROSECOND), owner=NULL
        WHERE id=?`, stPending, retry+1, backoff.Microseconds(), id)
	return err
}
func (s *Store) MarkFailed(ctx context.Context, id int64) error {
	_, err := s.Conn.ExecCtx(ctx, `UPDATE txn_message SET status=?, updated_at=NOW(6), owner=NULL WHERE id=?`, stFailed, id)
	return err
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
