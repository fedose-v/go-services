package temporal

import (
	"context"

	"go.temporal.io/sdk/client"

	"inventory/pkg/inventory/domain/model"
	"inventory/pkg/inventory/infrastructure/temporal/workflows"
)

const TaskQueue = "inventory_task_queue"

type WorkflowService interface {
	RunProductCreatedWorkflow(ctx context.Context, id string, event model.ProductCreated) error
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

func (s *workflowService) RunProductCreatedWorkflow(ctx context.Context, id string, event model.ProductCreated) error {
	_, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        id,
			TaskQueue: TaskQueue,
		},
		workflows.UserUpdatedWorkflow, event,
	)
	return err
}
