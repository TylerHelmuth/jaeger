// Copyright (c) 2019 The Jaeger Authors.
// Copyright (c) 2017 Uber Technologies, Inc.
// SPDX-License-Identifier: Apache-2.0

package adjuster_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/model/adjuster"
)

func TestSequences(t *testing.T) {
	// mock adjuster that increments span ID
	adj := adjuster.Func(func(trace *model.Trace) (*model.Trace, error) {
		trace.Spans[0].SpanID++
		return trace, nil
	})

	adjErr := errors.New("mock adjuster error")
	failingAdj := adjuster.Func(func(trace *model.Trace) (*model.Trace, error) {
		return trace, adjErr
	})

	testCases := []struct {
		adjuster   adjuster.Adjuster
		err        string
		lastSpanID int
	}{
		{
			adjuster:   adjuster.Sequence(adj, failingAdj, adj, failingAdj),
			err:        fmt.Sprintf("%s\n%s", adjErr, adjErr),
			lastSpanID: 2,
		},
		{
			adjuster:   adjuster.FailFastSequence(adj, failingAdj, adj, failingAdj),
			err:        adjErr.Error(),
			lastSpanID: 1,
		},
	}

	for _, testCase := range testCases {
		span := &model.Span{}
		trace := model.Trace{Spans: []*model.Span{span}}

		adjTrace, err := testCase.adjuster.Adjust(&trace)

		assert.Equal(t, span, adjTrace.Spans[0], "same trace & span returned")
		assert.EqualValues(t, testCase.lastSpanID, span.SpanID, "expect span ID to be incremented")
		require.EqualError(t, err, testCase.err)
	}
}
