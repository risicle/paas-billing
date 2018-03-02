// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/alphagov/paas-billing/collector"
)

type FakeEventFetcher struct {
	NameStub        func() string
	nameMutex       sync.RWMutex
	nameArgsForCall []struct{}
	nameReturns     struct {
		result1 string
	}
	nameReturnsOnCall map[int]struct {
		result1 string
	}
	FetchEventsStub        func(logger lager.Logger, fetchLimit int, recordMinAge time.Duration) (int, error)
	fetchEventsMutex       sync.RWMutex
	fetchEventsArgsForCall []struct {
		logger       lager.Logger
		fetchLimit   int
		recordMinAge time.Duration
	}
	fetchEventsReturns struct {
		result1 int
		result2 error
	}
	fetchEventsReturnsOnCall map[int]struct {
		result1 int
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeEventFetcher) Name() string {
	fake.nameMutex.Lock()
	ret, specificReturn := fake.nameReturnsOnCall[len(fake.nameArgsForCall)]
	fake.nameArgsForCall = append(fake.nameArgsForCall, struct{}{})
	fake.recordInvocation("Name", []interface{}{})
	fake.nameMutex.Unlock()
	if fake.NameStub != nil {
		return fake.NameStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.nameReturns.result1
}

func (fake *FakeEventFetcher) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *FakeEventFetcher) NameReturns(result1 string) {
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeEventFetcher) NameReturnsOnCall(i int, result1 string) {
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeEventFetcher) FetchEvents(logger lager.Logger, fetchLimit int, recordMinAge time.Duration) (int, error) {
	fake.fetchEventsMutex.Lock()
	ret, specificReturn := fake.fetchEventsReturnsOnCall[len(fake.fetchEventsArgsForCall)]
	fake.fetchEventsArgsForCall = append(fake.fetchEventsArgsForCall, struct {
		logger       lager.Logger
		fetchLimit   int
		recordMinAge time.Duration
	}{logger, fetchLimit, recordMinAge})
	fake.recordInvocation("FetchEvents", []interface{}{logger, fetchLimit, recordMinAge})
	fake.fetchEventsMutex.Unlock()
	if fake.FetchEventsStub != nil {
		return fake.FetchEventsStub(logger, fetchLimit, recordMinAge)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.fetchEventsReturns.result1, fake.fetchEventsReturns.result2
}

func (fake *FakeEventFetcher) FetchEventsCallCount() int {
	fake.fetchEventsMutex.RLock()
	defer fake.fetchEventsMutex.RUnlock()
	return len(fake.fetchEventsArgsForCall)
}

func (fake *FakeEventFetcher) FetchEventsArgsForCall(i int) (lager.Logger, int, time.Duration) {
	fake.fetchEventsMutex.RLock()
	defer fake.fetchEventsMutex.RUnlock()
	return fake.fetchEventsArgsForCall[i].logger, fake.fetchEventsArgsForCall[i].fetchLimit, fake.fetchEventsArgsForCall[i].recordMinAge
}

func (fake *FakeEventFetcher) FetchEventsReturns(result1 int, result2 error) {
	fake.FetchEventsStub = nil
	fake.fetchEventsReturns = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeEventFetcher) FetchEventsReturnsOnCall(i int, result1 int, result2 error) {
	fake.FetchEventsStub = nil
	if fake.fetchEventsReturnsOnCall == nil {
		fake.fetchEventsReturnsOnCall = make(map[int]struct {
			result1 int
			result2 error
		})
	}
	fake.fetchEventsReturnsOnCall[i] = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeEventFetcher) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	fake.fetchEventsMutex.RLock()
	defer fake.fetchEventsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeEventFetcher) recordInvocation(key string, args []interface{}) {
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

var _ collector.EventFetcher = new(FakeEventFetcher)
