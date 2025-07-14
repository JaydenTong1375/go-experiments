package utils

import "sort"

type List[T comparable] struct {
	head, tail *element[T]
}

type element[T any] struct {
	next *element[T]
	val  T
}

// Append value to the list
func (lst *List[T]) Append(val T) {
	e := &element[T]{val: val}
	if lst.tail == nil {
		lst.head = e
		lst.tail = e
	} else {
		lst.tail.next = e
		lst.tail = e
	}
}

// SortBy accepts a custom less function for sorting
func (lst *List[T]) SortBy(less func(a, b T) bool) {
	// Step 1: Convert list to slice
	var slice []T
	for cur := lst.head; cur != nil; cur = cur.next {
		slice = append(slice, cur.val)
	}

	// Step 2: Sort the slice
	sort.Slice(slice, func(i, j int) bool {
		return less(slice[i], slice[j])
	})

	// Step 3: Rebuild the linked list from sorted slice
	lst.head, lst.tail = nil, nil
	for _, val := range slice {
		lst.Append(val)
	}
}

func (lst *List[T]) Push(v T) {
	if lst.tail == nil {
		lst.head = &element[T]{val: v}
		lst.tail = lst.head
	} else {
		lst.tail.next = &element[T]{val: v}
		lst.tail = lst.tail.next
	}
}

func (lst *List[T]) IndexOf(v T) int {
	current := lst.head
	index := 0

	for current != nil {
		if current.val == v {
			return index
		}
		current = current.next
		index++
	}

	return -1
}

func (lst *List[T]) AllElements() []T {
	var elems []T
	for e := lst.head; e != nil; e = e.next {
		elems = append(elems, e.val)
	}
	return elems
}

func (lst *List[T]) Length() int {
	count := 0
	current := lst.head
	for current != nil {
		count++
		current = current.next
	}
	return count
}
