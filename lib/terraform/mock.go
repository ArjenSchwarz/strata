package terraform

import (
	"context"
)

// MockExecutor is a mock implementation of TerraformExecutor for testing
type MockExecutor struct {
	CheckInstallationFunc func(ctx context.Context) error
	GetVersionFunc        func(ctx context.Context) (string, error)
	PlanFunc              func(ctx context.Context, args []string) (string, error)
	ApplyFunc             func(ctx context.Context, planFile string, args []string) error
	DetectBackendFunc     func(ctx context.Context) (*BackendConfig, error)
	ValidateBackendFunc   func(ctx context.Context, config *BackendConfig) error
}

func (m *MockExecutor) CheckInstallation(ctx context.Context) error {
	if m.CheckInstallationFunc != nil {
		return m.CheckInstallationFunc(ctx)
	}
	return nil
}

func (m *MockExecutor) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}
	return "1.6.0", nil
}

func (m *MockExecutor) Plan(ctx context.Context, args []string) (string, error) {
	if m.PlanFunc != nil {
		return m.PlanFunc(ctx, args)
	}
	return "/tmp/test.tfplan", nil
}

func (m *MockExecutor) Apply(ctx context.Context, planFile string, args []string) error {
	if m.ApplyFunc != nil {
		return m.ApplyFunc(ctx, planFile, args)
	}
	return nil
}

func (m *MockExecutor) DetectBackend(ctx context.Context) (*BackendConfig, error) {
	if m.DetectBackendFunc != nil {
		return m.DetectBackendFunc(ctx)
	}
	return &BackendConfig{Type: "local"}, nil
}

func (m *MockExecutor) ValidateBackend(ctx context.Context, config *BackendConfig) error {
	if m.ValidateBackendFunc != nil {
		return m.ValidateBackendFunc(ctx, config)
	}
	return nil
}
