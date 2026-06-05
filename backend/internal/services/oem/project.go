package oem

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"errors"
	"strings"
	"time"

	"candypro/api/internal/pkg/crypto"
)

// ErrOEMStatusConflict signals that an OEM project status update lost a race
// against a concurrent writer — the row's current status no longer matches
// the version we validated against (R2 A-4).
var ErrOEMStatusConflict = errors.New("oem project status changed concurrently")

// ErrOEMProjectConflict signals that a general project mutation lost an
// optimistic-lock race.
var ErrOEMProjectConflict = errors.New("oem project changed concurrently")

// ErrOEMSampleNotFound is returned when a sample id cannot be located on a project.
var ErrOEMSampleNotFound = errors.New("oem sample not found")

// ErrOEMSampleStatusInvalid is returned for an unrecognised sample status.
var ErrOEMSampleStatusInvalid = errors.New("oem sample status invalid")

type projectRepository interface {
	FindAll(ctx context.Context, page, limit int, search, status string) ([]modelsProduct.OEMProject, int64, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsProduct.OEMProject, int64, error)
	FindByID(ctx context.Context, id string) (*modelsProduct.OEMProject, error)
	Create(ctx context.Context, project *modelsProduct.OEMProject) error
	Update(ctx context.Context, project *modelsProduct.OEMProject) error
	UpdateIfVersionMatches(ctx context.Context, project *modelsProduct.OEMProject, expectedVersion uint) (int64, error)
	UpdateStatusIfMatches(ctx context.Context, id, expected, target string) (int64, error)
	LinkOrderIfEmpty(ctx context.Context, id, orderID string) (int64, error)
	CreateLinkedOrderIfEmpty(ctx context.Context, id string, order *modelsOrder.Order) (bool, error)
	Delete(ctx context.Context, id string) error
}

type ProjectService struct {
	repo projectRepository
}

func NewProjectService(repo projectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetProjects(ctx context.Context, page, limit int, search, status string) ([]modelsProduct.OEMProject, int64, error) {
	return s.repo.FindAll(ctx, page, limit, search, status)
}

func (s *ProjectService) GetUserProjects(ctx context.Context, userID string, page, limit int) ([]modelsProduct.OEMProject, int64, error) {
	return s.repo.FindByUserID(ctx, userID, page, limit)
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (*modelsProduct.OEMProject, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProjectService) CreateProject(ctx context.Context, project *modelsProduct.OEMProject) error {
	if project.Status == "" {
		project.Status = "inquiry"
	}
	return s.repo.Create(ctx, project)
}

func (s *ProjectService) UpdateProject(ctx context.Context, project *modelsProduct.OEMProject) error {
	expectedVersion := normalizedVersion(project.Version)
	rows, err := s.repo.UpdateIfVersionMatches(ctx, project, expectedVersion)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOEMProjectConflict
	}
	project.Version = expectedVersion + 1
	return nil
}

func (s *ProjectService) UpdateProjectStatus(ctx context.Context, id string, status string) error {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// R4-10: Validate status transition
	if err := modelsProduct.ValidateOEMStatusTransition(project.Status, status); err != nil {
		return err
	}
	// R2 A-4: previously this read the row, validated in-memory, then issued a
	// blind Save() — two concurrent admins could each see the same source state
	// and silently overwrite each other. The conditional UPDATE makes the
	// original status part of the WHERE clause; if a competing writer changed
	// it first, RowsAffected is 0 and we surface the conflict.
	rows, uerr := s.repo.UpdateStatusIfMatches(ctx, id, project.Status, status)
	if uerr != nil {
		return uerr
	}
	if rows == 0 {
		return ErrOEMStatusConflict
	}
	return nil
}

// AddSample appends a new sample (status "requested") to the project and
// returns the created sample. This makes the "sampling" stage actionable
// (P0.2 / G-OEM-1) instead of a status string with no backing data.
func (s *ProjectService) AddSample(ctx context.Context, projectID, name string) (*modelsProduct.OEMSample, error) {
	sample := modelsProduct.OEMSample{
		ID:     crypto.GenerateID(),
		Name:   strings.TrimSpace(name),
		Status: modelsProduct.OEMSampleStatusRequested,
	}
	if err := s.mutateProject(ctx, projectID, func(project *modelsProduct.OEMProject) error {
		project.Samples = append([]modelsProduct.OEMSample(project.Samples), sample)
		return nil
	}); err != nil {
		return nil, err
	}
	return &sample, nil
}

// UpdateSampleStatus advances a sample's lifecycle (requested → shipped →
// received → approved/rejected), stamping sent/received timestamps and
// recording optional feedback. Returns the updated sample.
func (s *ProjectService) UpdateSampleStatus(ctx context.Context, projectID, sampleID, status, feedback string) (*modelsProduct.OEMSample, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if !modelsProduct.IsValidOEMSampleStatus(status) {
		return nil, ErrOEMSampleStatusInvalid
	}
	var updated modelsProduct.OEMSample
	err := s.mutateProject(ctx, projectID, func(project *modelsProduct.OEMProject) error {
		samples := []modelsProduct.OEMSample(project.Samples)
		idx := -1
		for i := range samples {
			if samples[i].ID == sampleID {
				idx = i
				break
			}
		}
		if idx == -1 {
			return ErrOEMSampleNotFound
		}
		now := time.Now()
		samples[idx].Status = status
		if feedback = strings.TrimSpace(feedback); feedback != "" {
			samples[idx].Feedback = feedback
		}
		switch status {
		case modelsProduct.OEMSampleStatusShipped:
			if samples[idx].SentAt == nil {
				samples[idx].SentAt = &now
			}
		case modelsProduct.OEMSampleStatusReceived,
			modelsProduct.OEMSampleStatusApproved,
			modelsProduct.OEMSampleStatusRejected:
			if samples[idx].ReceivedAt == nil {
				samples[idx].ReceivedAt = &now
			}
		}
		project.Samples = samples
		updated = samples[idx]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// LinkOrder records the converted order id on the project using a conditional
// update so a project can only be converted once (P0.2 / G-OEM-2). Returns
// false when the project already has an order linked.
func (s *ProjectService) LinkOrder(ctx context.Context, projectID, orderID string) (bool, error) {
	rows, err := s.repo.LinkOrderIfEmpty(ctx, projectID, orderID)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *ProjectService) CreateLinkedOrder(ctx context.Context, projectID string, order *modelsOrder.Order) (bool, error) {
	return s.repo.CreateLinkedOrderIfEmpty(ctx, projectID, order)
}

const projectMutationAttempts = 3

func (s *ProjectService) mutateProject(ctx context.Context, projectID string, mutate func(*modelsProduct.OEMProject) error) error {
	for attempt := 0; attempt < projectMutationAttempts; attempt++ {
		project, err := s.repo.FindByID(ctx, projectID)
		if err != nil {
			return err
		}
		if err := mutate(project); err != nil {
			return err
		}
		if err := s.UpdateProject(ctx, project); !errors.Is(err, ErrOEMProjectConflict) {
			return err
		}
	}
	return ErrOEMProjectConflict
}

func normalizedVersion(version uint) uint {
	if version == 0 {
		return 1
	}
	return version
}
