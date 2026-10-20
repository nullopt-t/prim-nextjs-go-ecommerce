package setting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type captureExecutor struct {
	args []interface{}
}

func (e *captureExecutor) Exec(_ context.Context, _ string, args ...interface{}) (pgconn.CommandTag, error) {
	e.args = args
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (e *captureExecutor) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (e *captureExecutor) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	return nil
}

func TestRepositoryUpsertEncodesValueAsJSON(t *testing.T) {
	exec := &captureExecutor{}
	repo := NewRepository()

	if err := repo.Upsert(context.Background(), exec, "storeName", "PRIM Superstore"); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if len(exec.args) != 2 {
		t.Fatalf("Upsert() args = %d, want 2", len(exec.args))
	}

	payload, ok := exec.args[1].([]byte)
	if !ok {
		t.Fatalf("Upsert() value type = %T, want []byte", exec.args[1])
	}
	if !json.Valid(payload) {
		t.Fatalf("Upsert() value is not valid JSON: %s", payload)
	}

	var got string
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got != "PRIM Superstore" {
		t.Fatalf("Upsert() value = %q, want %q", got, "PRIM Superstore")
	}
}
