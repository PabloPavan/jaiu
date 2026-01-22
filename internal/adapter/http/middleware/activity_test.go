package middleware

import (
	"context"
	"testing"
)

// Testa que o contexto retorna a mesma instancia de atividade criada.
func TestUserActivityContext(t *testing.T) {
	ctx, activity := withUserActivity(context.Background())

	got, ok := userActivityFromContext(ctx)
	if !ok {
		t.Fatal("expected activity in context")
	}
	if got != activity {
		t.Fatal("expected same activity pointer from context")
	}
}
