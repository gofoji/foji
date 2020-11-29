package stringlist

type StringMap map[string]string

func (t StringMap) IsEmpty() bool {
	return len(t) == 0
}

func (t StringMap) Values() Strings {
	result := Strings{}

	for _, v := range t {
		result = append(result, v)
	}

	return result
}
