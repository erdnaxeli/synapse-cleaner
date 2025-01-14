package synapsecleaner

// Return the elements in a not present in b.
func DiffSlices[S ~[]E, E comparable](a S, b S) S {
	identity := func(e E) E { return e }
	return DiffSlicesFunc(
		a,
		b,
		identity,
		identity,
	)
}

// Return the elements in a not present in b.
//
// The methodes keyA() and keyB() will be called for each element of respectively
// slices a and b. They must return some value that can compared.
func DiffSlicesFunc[S ~[]E, T ~[]F, E, F, G comparable](a S, b T, keyA func(E) G, keyB func(F) G) S {
	indexed := make(map[G]struct{}, len(b))
	for _, e := range b {
		indexed[keyB(e)] = struct{}{}
	}

	result := make(S, 0, max(len(a)-len(indexed), 0))
	for _, e := range a {
		if _, ok := indexed[keyA(e)]; !ok {
			result = append(result, e)
		}
	}

	return result
}
