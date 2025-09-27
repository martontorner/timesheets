package arrays

func Map[T, U any](in []T, f func(T) U) []U {
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func MapE[T, U any](in []T, f func(T) (U, error)) ([]U, error) {
	out := make([]U, len(in))
	for i, v := range in {
		u, err := f(v)
		if err != nil {
			return nil, err
		}

		out[i] = u
	}
	return out, nil
}

func Filter[T any](in []T, f func(T) bool) []T {
	out := make([]T, 0, len(in))
	for _, v := range in {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}
