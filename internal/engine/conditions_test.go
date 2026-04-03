package engine

import (
	"testing"

	"github.com/schigh/svctmpl/internal/genome"
)

func TestEvaluateConditions_BooleanAxisComposeTrue(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Router:    "chi",
		Database:  "postgres",
		Compose:   true,
	}
	if !EvaluateConditions([]string{"compose"}, choices) {
		t.Error("bare 'compose' should pass when Compose=true")
	}
}

func TestEvaluateConditions_BooleanAxisComposeFalse(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Router:    "chi",
		Database:  "postgres",
		Compose:   false,
	}
	if EvaluateConditions([]string{"compose"}, choices) {
		t.Error("bare 'compose' should fail when Compose=false")
	}
}

func TestEvaluateConditions_BooleanAxisK8sExplicitTrue(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Router:    "chi",
		Database:  "postgres",
		K8s:       true,
	}
	if !EvaluateConditions([]string{"k8s:true"}, choices) {
		t.Error("'k8s:true' should pass when K8s=true")
	}
}

func TestEvaluateConditions_BooleanAxisK8sExplicitFalse(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Router:    "chi",
		Database:  "postgres",
		K8s:       false,
	}
	if EvaluateConditions([]string{"k8s:true"}, choices) {
		t.Error("'k8s:true' should fail when K8s=false")
	}
}

func TestEvaluateConditions_BooleanAxisTiltBare(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Tilt:      true,
	}
	if !EvaluateConditions([]string{"tilt"}, choices) {
		t.Error("bare 'tilt' should pass when Tilt=true")
	}
}

func TestEvaluateConditions_MixedBooleanAndString(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Router:    "chi",
		Database:  "postgres",
		Compose:   true,
	}
	// AND logic: database is not none AND compose is true
	if !EvaluateConditions([]string{"database", "compose"}, choices) {
		t.Error("expected [database, compose] to pass with database=postgres and compose=true")
	}
}

func TestEvaluateConditions_MixedBooleanAndString_OneFails(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Router:    "chi",
		Database:  "postgres",
		Compose:   false,
	}
	// AND logic: database passes but compose fails
	if EvaluateConditions([]string{"database", "compose"}, choices) {
		t.Error("expected [database, compose] to fail when compose=false")
	}
}

func TestEvaluateConditions_KeyValueExactMatch(t *testing.T) {
	choices := &genome.Choices{
		Transport:     "http",
		Router:        "chi",
		Database:      "postgres",
		DBTooling:     "sqlc",
		Observability: "otel-full",
	}
	if !EvaluateConditions([]string{"db_tooling:sqlc"}, choices) {
		t.Error("'db_tooling:sqlc' should pass when DBTooling=sqlc")
	}
	if EvaluateConditions([]string{"db_tooling:none"}, choices) {
		t.Error("'db_tooling:none' should fail when DBTooling=sqlc")
	}
}

func TestEvaluateConditions_UnknownAxis(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
	}
	// Unknown bare key should fail (not in the map)
	if EvaluateConditions([]string{"nonexistent"}, choices) {
		t.Error("unknown bare axis should fail")
	}
	// Unknown key:value should fail
	if EvaluateConditions([]string{"nonexistent:value"}, choices) {
		t.Error("unknown key:value axis should fail")
	}
}

func TestEvaluateConditions_DatabaseNone(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Router:    "chi",
		Database:  "none",
	}
	if EvaluateConditions([]string{"database"}, choices) {
		t.Error("bare 'database' should fail when database=none")
	}
}
