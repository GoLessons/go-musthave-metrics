package container

import "fmt"

type Container interface {
	Get(id string) (entry *any, err error)
}

func GetService[T any](container Container, id string) (*T, error) {
	service, err := container.Get(id)
	if err != nil {
		return nil, err
	}

	result := (*service).(*T)
	return result, nil
}

type Factory[T any] func(container Container) (T, error)

type ContainerError struct {
	Msg string
}

func Error(format string, a ...any) error {
	return &ContainerError{
		Msg: fmt.Sprintf(format, a...),
	}
}

func (e ContainerError) Error() string {
	return fmt.Sprintf("container error: %s", e.Msg)
}

type simpleContainer struct {
	services  map[string]*any
	factories map[string]Factory[any]
}

func NewSimpleContainer(services map[string]any) simpleContainer {
	container := simpleContainer{
		services:  map[string]*any{},
		factories: map[string]Factory[any]{},
	}
	for name, service := range services {
		if _, ok := service.(Factory[any]); ok {
			container.RegisterFactory(name, service.(Factory[any]))
		} else {
			container.RegisterService(name, &service)
		}
	}

	return container
}

func (container simpleContainer) Get(id string) (*any, error) {
	fmt.Printf("Get %s in %v\n", id, container.services)

	entry, has := container.services[id]
	if !has {
		fmt.Printf("Not has %s in %v\n", id, container.services)

		var err error
		entry, err = container.create(id)
		if err != nil {
			return nil, err
		}

		container.services[id] = entry
		return entry, nil
	}

	return entry, nil
}

func (container simpleContainer) RegisterFactory(id string, factory Factory[any]) {
	container.factories[id] = factory
}

func (container simpleContainer) RegisterService(id string, service *any) {
	container.services[id] = service
}

func (container simpleContainer) create(id string) (*any, error) {
	factory, ok := container.factories[id]
	if !ok {
		return nil, Error("factory for '%s' not found", id)
	}

	service, err := factory(container)
	return &service, err
}
