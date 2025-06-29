// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package queuev2

import (
	"github.com/uber/cadence/common/persistence"
)

type Range struct {
	InclusiveMinTaskKey persistence.HistoryTaskKey
	ExclusiveMaxTaskKey persistence.HistoryTaskKey
}

func (r *Range) IsEmpty() bool {
	return r.InclusiveMinTaskKey.Compare(r.ExclusiveMaxTaskKey) >= 0
}

func (r *Range) Contains(taskKey persistence.HistoryTaskKey) bool {
	return taskKey.Compare(r.InclusiveMinTaskKey) >= 0 && taskKey.Compare(r.ExclusiveMaxTaskKey) < 0
}

func (r *Range) ContainsRange(other Range) bool {
	return r.InclusiveMinTaskKey.Compare(other.InclusiveMinTaskKey) <= 0 && r.ExclusiveMaxTaskKey.Compare(other.ExclusiveMaxTaskKey) >= 0
}

func (r *Range) CanMerge(other Range) bool {
	return r.InclusiveMinTaskKey.Compare(other.ExclusiveMaxTaskKey) <= 0 && r.ExclusiveMaxTaskKey.Compare(other.InclusiveMinTaskKey) >= 0
}

func (r *Range) CanSplitByTaskKey(taskKey persistence.HistoryTaskKey) bool {
	return taskKey.Compare(r.InclusiveMinTaskKey) > 0 && taskKey.Compare(r.ExclusiveMaxTaskKey) < 0
}
