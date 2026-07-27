//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type healthProbeRepoStub struct {
	active []Proxy
	marked []int64
}

func (r *healthProbeRepoStub) Create(context.Context, *Proxy) error { return nil }
func (r *healthProbeRepoStub) GetByID(context.Context, int64) (*Proxy, error) {
	return nil, ErrProxyNotFound
}
func (r *healthProbeRepoStub) ListByIDs(context.Context, []int64) ([]Proxy, error) { return nil, nil }
func (r *healthProbeRepoStub) Update(context.Context, *Proxy) error                 { return nil }
func (r *healthProbeRepoStub) Delete(context.Context, int64) error                  { return nil }
func (r *healthProbeRepoStub) List(context.Context, pagination.PaginationParams) ([]Proxy, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *healthProbeRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]Proxy, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *healthProbeRepoStub) ListWithFiltersAndAccountCount(context.Context, pagination.PaginationParams, string, string, string) ([]ProxyWithAccountCount, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *healthProbeRepoStub) ListActive(context.Context) ([]Proxy, error) {
	out := make([]Proxy, len(r.active))
	copy(out, r.active)
	return out, nil
}
func (r *healthProbeRepoStub) ListActiveWithAccountCount(context.Context) ([]ProxyWithAccountCount, error) {
	return nil, nil
}
func (r *healthProbeRepoStub) ExistsByHostPortAuth(context.Context, string, int, string, string) (bool, error) {
	return false, nil
}
func (r *healthProbeRepoStub) CountAccountsByProxyID(context.Context, int64) (int64, error) {
	return 0, nil
}
func (r *healthProbeRepoStub) ListAccountSummariesByProxyID(context.Context, int64) ([]ProxyAccountSummary, error) {
	return nil, nil
}
func (r *healthProbeRepoStub) SweepExpiredProxies(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (r *healthProbeRepoStub) MarkProxyUnhealthyAndReroute(_ context.Context, proxyID int64, _ time.Time) (int64, error) {
	r.marked = append(r.marked, proxyID)
	return 1, nil
}
func (r *healthProbeRepoStub) ListAllForFallback(context.Context) ([]Proxy, error) { return nil, nil }
func (r *healthProbeRepoStub) CountExpired(context.Context) (int64, error)         { return 0, nil }
func (r *healthProbeRepoStub) CountExpiringSoon(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type healthProberStub struct {
	// map proxy URL -> latency/error
	latency map[string]int64
	errs    map[string]error
}

func (p *healthProberStub) ProbeProxy(_ context.Context, proxyURL string) (*ProxyExitInfo, int64, error) {
	if err, ok := p.errs[proxyURL]; ok && err != nil {
		return nil, 0, err
	}
	if latency, ok := p.latency[proxyURL]; ok {
		return &ProxyExitInfo{IP: "1.1.1.1"}, latency, nil
	}
	return &ProxyExitInfo{IP: "1.1.1.1"}, 20, nil
}

func TestProxyExpiryService_ProbeAndFailoverUnhealthyProxies(t *testing.T) {
	dead := Proxy{ID: 1, Name: "dead", Protocol: "http", Host: "127.0.0.1", Port: 1, Status: StatusActive}
	alive := Proxy{ID: 2, Name: "alive", Protocol: "http", Host: "127.0.0.1", Port: 2, Status: StatusActive}
	repo := &healthProbeRepoStub{active: []Proxy{dead, alive}}
	prober := &healthProberStub{
		latency: map[string]int64{
			alive.URL(): 35,
			dead.URL():  0,
		},
		errs: map[string]error{
			dead.URL(): errors.New("proxy connection failed"),
		},
	}

	svc := NewProxyExpiryService(repo, time.Minute)
	svc.SetProxyProber(prober)

	changed, err := svc.probeAndFailoverUnhealthyProxies(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), changed)
	require.Equal(t, []int64{1}, repo.marked)
}

func TestProxyExpiryService_ProbeNoLatencyAlsoFailsOver(t *testing.T) {
	noLatency := Proxy{ID: 9, Name: "zero", Protocol: "http", Host: "127.0.0.1", Port: 9, Status: StatusActive}
	repo := &healthProbeRepoStub{active: []Proxy{noLatency}}
	prober := &healthProberStub{
		latency: map[string]int64{noLatency.URL(): 0},
	}
	svc := NewProxyExpiryService(repo, time.Minute)
	svc.SetProxyProber(prober)

	changed, err := svc.probeAndFailoverUnhealthyProxies(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), changed)
	require.Equal(t, []int64{9}, repo.marked)
}
