package pipeline

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func __test_pipeline1(t *testing.T) (*Pipeline, error) {
	return SetPipeline("pipeline_1", // 1 diag event for the start
		AnyAgent(),
		Stages("stages_1", // 1 diag for the stages
			Stage("stage_1", // 1 diag for the stage
				SH("echo", "bonjour"),
				SH("echo", "bonjour"),
				SH("echo", "bonjour"),
			), // 1 diag event at the end of stage_1
			Stage("stage_2", // 1 diag for the stage
				SH("echo", "bonjour"),
				SH("echo", "bonjour"),
				SH("echo", "bonjour"),
			), // 1 diag event at the end of stages_1
		), // 1 diag event at the end
	)
}

func __test_pipeline2(t *testing.T) (*Pipeline, error) {
	return SetPipeline("pipeline_1", // 1 diag event for the start
		AnyAgent(),
		Stages("stages_1", // 1 diag for the stages
			Stage("stage_1", // 1 diag for the stage
				Exec(func(p *Pipeline, ctx context.Context) error { return nil }),
			), // 1 diag event at the end of stage_1
			Stage("stage_2", // 1 diag for the stage
				Exec(func(p *Pipeline, ctx context.Context) error { return errors.New("test") }), // 1 diag for error
			), // 1 diag event at the end of stage_2
			Stage("stage_3", // no diag here
				Exec(func(p *Pipeline, ctx context.Context) error { return errors.New("test") }),
			), // no diag here
		), // diag event at the end
	) // diag event at the end
}

func __test__getPipelineDiagnostics(t *testing.T, f func(t *testing.T) (*Pipeline, error)) *Pipeline {

	p1, err := f(t)

	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}

	err = p1.ExecutePipeline(context.Background())

	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	// Only keep INFO level logs to keep it sane
	p1.Diagnostic = p1.Diagnostic.FilterBasedOnImportance(INFO)

	return p1

}

func TestDiagnostics(t *testing.T) {
	p1 := __test__getPipelineDiagnostics(t, __test_pipeline1)
	expected := "pipeline_1"
	actual := p1.Diagnostic.Label
	if actual != expected {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}

	expected = "3"
	actual = fmt.Sprintf("%d", len(p1.Diagnostic.Events))

	if actual != expected {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}

	switch p1.Diagnostic.Events[0].(type) {
	case *Diagnostic:
		t.Fatalf("first diag event of wrong type")
	case *DiagnosticEvent:
		break
	}
	switch p1.Diagnostic.Events[1].(type) {
	case *Diagnostic:
		break
	case *DiagnosticEvent:
		t.Fatalf("second diag event of wrong type")
	}
	switch p1.Diagnostic.Events[2].(type) {
	case *Diagnostic:
		t.Fatalf("first diag event of wrong type")
	case *DiagnosticEvent:
		break
	}

	diag := p1.Diagnostic.Events[1].(*Diagnostic)

	expected = "4"
	actual = fmt.Sprintf("%d", len(diag.Events))

	if actual != expected {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}
	switch diag.Events[0].(type) {
	case *Diagnostic:
		t.Fatalf("first diag event of wrong type")
	case *DiagnosticEvent:
		break
	}
	switch diag.Events[1].(type) {
	case *Diagnostic:
		break
	case *DiagnosticEvent:
		t.Fatalf("second diag event of wrong type")
	}
	switch diag.Events[2].(type) {
	case *Diagnostic:
		break
	case *DiagnosticEvent:
		t.Fatalf("second diag event of wrong type")
	}
	switch diag.Events[3].(type) {
	case *Diagnostic:
		t.Fatalf("first diag event of wrong type")
	case *DiagnosticEvent:
		break
	}

	diag = diag.Events[1].(*Diagnostic)

	expected = "2"
	actual = fmt.Sprintf("%d", len(diag.Events))

	if actual != expected {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}
	switch diag.Events[0].(type) {
	case *Diagnostic:
		t.Fatalf("first diag event of wrong type")
	case *DiagnosticEvent:
		break
	}
	switch diag.Events[1].(type) {
	case *Diagnostic:
		t.Fatalf("first diag event of wrong type")
	case *DiagnosticEvent:
		break
	}
}

func TestDiagErrorRecovery(t *testing.T) {
	p1 := __test__getPipelineDiagnostics(t, __test_pipeline2)

	expected := "pipeline_1"
	actual := p1.Diagnostic.Label
	if actual != expected {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}

	// TODO: finish test
}
