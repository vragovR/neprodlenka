package names

import (
	"cmp"
	"slices"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func Sort[T any](items []T, name func(T) (last, first string)) {
	col := collate.New(language.Russian)
	slices.SortStableFunc(items, func(a, b T) int {
		aLast, aFirst := name(a)
		bLast, bFirst := name(b)
		return cmp.Or(col.CompareString(aLast, bLast), col.CompareString(aFirst, bFirst))
	})
}
