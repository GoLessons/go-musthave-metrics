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
		container.registerService(name, &service)
	}

	return container
}

func SimpleRegisterFactory[T any](container *simpleContainer, id string, factory Factory[T]) {
	container.registerFactory(id, func(c Container) (any, error) {
		return factory(c)
	})
}

func (container simpleContainer) Get(id string) (*any, error) {
	entry, has := container.services[id]
	if !has {
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

func (container *simpleContainer) registerFactory(id string, factory Factory[any]) {
	container.factories[id] = factory
}

func (container simpleContainer) registerService(id string, service *any) {
	container.services[id] = service
}

func (container simpleContainer) Alias(id string, as string) error {
	service, err := container.Get(id)
	if err != nil {
		return err
	}

	container.registerService(as, service)
	return nil
}

func (container simpleContainer) create(id string) (*any, error) {
	factory, ok := container.factories[id]
	if !ok {
		return nil, Error("factory for '%s' not found", id)
	}

	service, err := factory(container)
	return &service, err
}
