// Copyright (c) 2020 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package idalloc

import (
	"errors"
	"sort"

	"github.com/golang-collections/collections/stack"

	"github.com/projectcalico/calico/libcalico-go/lib/set"
)

type IndexRange struct {
	Min, Max int
}

// ByMaxIndex sorts collections of IndexRange structs in order of their starting/lower index
type ByMaxIndex []IndexRange

// Len is the number of indexranges in the collection
func (i ByMaxIndex) Len() int { return len(i) }

// Less reports whether the element with index a
// must sort before the element with index b.
func (i ByMaxIndex) Less(a, b int) bool { return i[a].Max < i[b].Max }

// Swap swaps the elements with indexes a and b.
func (i ByMaxIndex) Swap(a, b int) { i[a], i[b] = i[b], i[a] }

type IndexAllocator struct {
	indexStack *stack.Stack
}

func NewIndexAllocator(indexRanges ...IndexRange) *IndexAllocator {
	// sort index ranges in descending order of their Max bound
	if len(indexRanges) > 1 {
		sort.Sort(sort.Reverse(ByMaxIndex(indexRanges)))
	}

	r := &IndexAllocator{
		indexStack: stack.New(),
	}

	var lowestIndex int
	for j, indexRange := range indexRanges {
		if j == 0 {
			// keep track of the most recent Min to prevent an overlapping range
			// from creating duplicate indices in the final index stack
			lowestIndex = indexRange.Max + 1
		}

		// Push in reverse order so that the lowest index will come out first.
		for i := indexRange.Max; i >= indexRange.Min; i-- {
			// skip overlapping range indices
			if i >= lowestIndex {
				continue
			}
			r.indexStack.Push(i)
			lowestIndex = i
		}
	}
	return r
}

func (r *IndexAllocator) GrabIndex() (int, error) {
	if r.indexStack.Len() == 0 {
		return 0, errors.New("no more indices available")
	}
	return r.indexStack.Pop().(int), nil
}

func (r *IndexAllocator) ReleaseIndex(index int) {
	r.indexStack.Push(index)
}

func (r *IndexAllocator) GrabAllRemainingIndices() set.Set {
	remainingIndices := set.New()
	idx, err := r.GrabIndex()
	for err == nil {
		remainingIndices.Add(idx)
		idx, err = r.GrabIndex()
	}
	return remainingIndices
}
