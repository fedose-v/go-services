package main

// TODO: добавить зависимости

func newDependencyContainer(
	_ *config,
	connContainer *connectionsContainer,
) (*dependencyContainer, error) {
	return &dependencyContainer{}, nil
}

type dependencyContainer struct {
}
