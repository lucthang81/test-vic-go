package utils

type ByInt64 []int64

func (a ByInt64) Len() int      { return len(a) }
func (a ByInt64) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByInt64) Less(i, j int) bool {
	return a[i] < a[j]
}

type ByInt []int

func (a ByInt) Len() int      { return len(a) }
func (a ByInt) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByInt) Less(i, j int) bool {
	return a[i] < a[j]
}
