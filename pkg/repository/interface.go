package gor

import "context"

type (
	// PageRequest selects a result window using a maximum item count and zero-based offset.
	PageRequest struct {
		// Limit is the maximum number of items to return.
		Limit int64
		// Offset is the number of matching items to skip.
		Offset int64
	}

	// Page contains one result window and the total number of matching items.
	Page[T any] struct {
		// Data contains the items in this result window.
		Data []T
		// Total is the total number of matching items across all windows.
		Total int64
	}

	// SearchSpec describes a repository search, including its requested result window.
	SearchSpec struct {
		PageRequest
	}

	// Repository defines common persistence operations for entities of type T keyed by ID.
	// Implementations should honor cancellation and deadlines from each method's context.
	Repository[T any, ID any] interface {
		// Create persists a new entity and returns its stored representation.
		Create(context.Context, T) (T, error)
		// Update persists changes to an existing entity and returns its stored representation.
		Update(context.Context, T) (T, error)
		// FindByID returns the entity identified by ID.
		FindByID(context.Context, ID) (T, error)
		// Delete removes the supplied entity.
		Delete(context.Context, T) error
		// FindAll returns a page of entities without additional search criteria.
		FindAll(context.Context, PageRequest) (Page[T], error)
		// Search returns a page of entities matching the supplied specification.
		Search(context.Context, SearchSpec) (Page[T], error)
	}
)
