package worker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"inventory/pkg/inventory/app/service"
	"inventory/pkg/inventory/infrastructure/temporal"
	"inventory/pkg/inventory/infrastructure/temporal/activity"
	"inventory/pkg/inventory/infrastructure/temporal/workflows"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	productService service.ProductService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})
	w.RegisterActivity(activity.NewInventoryServiceActivities(productService))
	w.RegisterWorkflow(workflows.UserUpdatedWorkflow)
	return w
}
