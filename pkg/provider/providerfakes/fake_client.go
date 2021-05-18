// Code generated by counterfeiter. DO NOT EDIT.
package providerfakes

import (
	"context"
	"sync"

	"github.com/homedepot/arcade/pkg/provider"
)

type FakeClient struct {
	TokenStub        func(context.Context) (string, error)
	tokenMutex       sync.RWMutex
	tokenArgsForCall []struct {
		arg1 context.Context
	}
	tokenReturns struct {
		result1 string
		result2 error
	}
	tokenReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClient) Token(arg1 context.Context) (string, error) {
	fake.tokenMutex.Lock()
	ret, specificReturn := fake.tokenReturnsOnCall[len(fake.tokenArgsForCall)]
	fake.tokenArgsForCall = append(fake.tokenArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.TokenStub
	fakeReturns := fake.tokenReturns
	fake.recordInvocation("Token", []interface{}{arg1})
	fake.tokenMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeClient) TokenCallCount() int {
	fake.tokenMutex.RLock()
	defer fake.tokenMutex.RUnlock()
	return len(fake.tokenArgsForCall)
}

func (fake *FakeClient) TokenCalls(stub func(context.Context) (string, error)) {
	fake.tokenMutex.Lock()
	defer fake.tokenMutex.Unlock()
	fake.TokenStub = stub
}

func (fake *FakeClient) TokenArgsForCall(i int) context.Context {
	fake.tokenMutex.RLock()
	defer fake.tokenMutex.RUnlock()
	argsForCall := fake.tokenArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeClient) TokenReturns(result1 string, result2 error) {
	fake.tokenMutex.Lock()
	defer fake.tokenMutex.Unlock()
	fake.TokenStub = nil
	fake.tokenReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeClient) TokenReturnsOnCall(i int, result1 string, result2 error) {
	fake.tokenMutex.Lock()
	defer fake.tokenMutex.Unlock()
	fake.TokenStub = nil
	if fake.tokenReturnsOnCall == nil {
		fake.tokenReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.tokenReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.tokenMutex.RLock()
	defer fake.tokenMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ provider.Client = new(FakeClient)
