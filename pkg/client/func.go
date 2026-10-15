package client

func WithElement[E ~[]T, T any](a E, e T) E {
	result := make(E, len(a), len(a)+1)
	copy(result, a)
	return append(result, e)
}
