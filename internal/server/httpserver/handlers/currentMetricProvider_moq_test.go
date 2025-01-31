// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package handlers

import (
	"context"
	"sync"
)

// Ensure, that CurrentMetricProviderMock does implement CurrentMetricProvider.
// If this is not the case, regenerate this file with moq.
var _ CurrentMetricProvider = &CurrentMetricProviderMock{}

// CurrentMetricProviderMock is a mock implementation of CurrentMetricProvider.
//
//	func TestSomethingThatUsesCurrentMetricProvider(t *testing.T) {
//
//		// make and configure a mocked CurrentMetricProvider
//		mockedCurrentMetricProvider := &CurrentMetricProviderMock{
//			GetValueFunc: func(contextMoqParam context.Context, s1 string, s2 string) (string, error) {
//				panic("mock out the GetValue method")
//			},
//		}
//
//		// use mockedCurrentMetricProvider in code that requires CurrentMetricProvider
//		// and then make assertions.
//
//	}
type CurrentMetricProviderMock struct {
	// GetValueFunc mocks the GetValue method.
	GetValueFunc func(contextMoqParam context.Context, s1 string, s2 string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetValue holds details about calls to the GetValue method.
		GetValue []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
		}
	}
	lockGetValue sync.RWMutex
}

// GetValue calls GetValueFunc.
func (mock *CurrentMetricProviderMock) GetValue(contextMoqParam context.Context, s1 string, s2 string) (string, error) {
	if mock.GetValueFunc == nil {
		panic("CurrentMetricProviderMock.GetValueFunc: method is nil but CurrentMetricProvider.GetValue was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		S1              string
		S2              string
	}{
		ContextMoqParam: contextMoqParam,
		S1:              s1,
		S2:              s2,
	}
	mock.lockGetValue.Lock()
	mock.calls.GetValue = append(mock.calls.GetValue, callInfo)
	mock.lockGetValue.Unlock()
	return mock.GetValueFunc(contextMoqParam, s1, s2)
}

// GetValueCalls gets all the calls that were made to GetValue.
// Check the length with:
//
//	len(mockedCurrentMetricProvider.GetValueCalls())
func (mock *CurrentMetricProviderMock) GetValueCalls() []struct {
	ContextMoqParam context.Context
	S1              string
	S2              string
} {
	var calls []struct {
		ContextMoqParam context.Context
		S1              string
		S2              string
	}
	mock.lockGetValue.RLock()
	calls = mock.calls.GetValue
	mock.lockGetValue.RUnlock()
	return calls
}
