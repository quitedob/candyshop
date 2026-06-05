package oem

import (
	"context"
	"sync"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
)

type projectTestRepository struct {
	mu           sync.Mutex
	project      *modelsProduct.OEMProject
	conflictNext bool
}

func (r *projectTestRepository) FindAll(context.Context, int, int, string, string) ([]modelsProduct.OEMProject, int64, error) {
	return nil, 0, nil
}

func (r *projectTestRepository) FindByUserID(context.Context, string, int, int) ([]modelsProduct.OEMProject, int64, error) {
	return nil, 0, nil
}

func (r *projectTestRepository) FindByID(context.Context, string) (*modelsProduct.OEMProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return cloneProject(r.project), nil
}

func (r *projectTestRepository) Create(_ context.Context, project *modelsProduct.OEMProject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.project = cloneProject(project)
	return nil
}

func (r *projectTestRepository) Update(_ context.Context, project *modelsProduct.OEMProject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.project = cloneProject(project)
	return nil
}

func (r *projectTestRepository) UpdateIfVersionMatches(_ context.Context, project *modelsProduct.OEMProject, expectedVersion uint) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conflictNext {
		r.conflictNext = false
		r.project.Samples = append(r.project.Samples, modelsProduct.OEMSample{ID: "concurrent", Name: "Concurrent sample"})
		r.project.Version++
		return 0, nil
	}
	if r.project.Version != expectedVersion {
		return 0, nil
	}
	r.project = cloneProject(project)
	r.project.Version = expectedVersion + 1
	return 1, nil
}

func (r *projectTestRepository) UpdateStatusIfMatches(_ context.Context, _, expected, target string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.project.Status != expected {
		return 0, nil
	}
	r.project.Status = target
	r.project.Version++
	return 1, nil
}

func (r *projectTestRepository) LinkOrderIfEmpty(_ context.Context, _, orderID string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.project.OrderID != nil {
		return 0, nil
	}
	r.project.OrderID = &orderID
	r.project.Version++
	return 1, nil
}

func (r *projectTestRepository) CreateLinkedOrderIfEmpty(_ context.Context, _ string, order *modelsOrder.Order) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.project.OrderID != nil {
		return false, nil
	}
	r.project.OrderID = &order.ID
	r.project.Version++
	return true, nil
}

func (r *projectTestRepository) Delete(context.Context, string) error { return nil }

func TestAddSampleRetriesAndMergesConcurrentSamples(t *testing.T) {
	repo := &projectTestRepository{
		project:      &modelsProduct.OEMProject{ID: "project-1", Version: 1},
		conflictNext: true,
	}
	service := NewProjectService(repo)

	sample, err := service.AddSample(context.Background(), "project-1", "Primary sample")
	if err != nil {
		t.Fatalf("AddSample() error = %v", err)
	}
	if sample.Name != "Primary sample" {
		t.Fatalf("sample name = %q, want Primary sample", sample.Name)
	}

	project, err := repo.FindByID(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if len(project.Samples) != 2 {
		t.Fatalf("sample count = %d, want 2", len(project.Samples))
	}
	if project.Samples[0].ID != "concurrent" || project.Samples[1].ID != sample.ID {
		t.Fatalf("samples = %#v, want concurrent sample followed by retried append", project.Samples)
	}
}

func TestCreateLinkedOrderOnlyLinksOnce(t *testing.T) {
	repo := &projectTestRepository{
		project: &modelsProduct.OEMProject{ID: "project-1", Version: 1},
	}
	service := NewProjectService(repo)

	linked, err := service.CreateLinkedOrder(context.Background(), "project-1", &modelsOrder.Order{ID: "order-1"})
	if err != nil {
		t.Fatalf("CreateLinkedOrder() error = %v", err)
	}
	if !linked {
		t.Fatal("CreateLinkedOrder() linked = false, want true")
	}

	linked, err = service.CreateLinkedOrder(context.Background(), "project-1", &modelsOrder.Order{ID: "order-2"})
	if err != nil {
		t.Fatalf("second CreateLinkedOrder() error = %v", err)
	}
	if linked {
		t.Fatal("second CreateLinkedOrder() linked = true, want false")
	}

	project, err := repo.FindByID(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if project.OrderID == nil || *project.OrderID != "order-1" {
		t.Fatalf("project.OrderID = %v, want order-1", project.OrderID)
	}
}

func cloneProject(project *modelsProduct.OEMProject) *modelsProduct.OEMProject {
	copy := *project
	copy.Samples = append(modelsProduct.OEMSampleArray(nil), project.Samples...)
	return &copy
}
