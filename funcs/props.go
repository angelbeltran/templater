package funcs

import "fmt"

// NewKVSProps is the implementation of the `props` template function.
func NewKVSProps(args ...any) (map[string]any, error) {
	if len(args)%2 == 1 {
		return nil, fmt.Errorf("the props function expects an even number of arguments, key-value pairs: received %d arguments", len(args))
	}

	props := make(map[string]any, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("props expected odd arguments to be key strings: argument %d was a %T", i+1, args[i])
		}

		props[k] = args[i+1]
	}

	return props, nil
}
