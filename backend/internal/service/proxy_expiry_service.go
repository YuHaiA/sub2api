package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	defaultProxyHealthProbeConcurrency = 3
	defaultProxyHealthProbeTimeout     = 12 * time.Second
)

var errProxyLatencyMissing = errors.New("proxy probe returned no latency")

// ProxyExpiryService 周期扫描到期/失活代理并把绑定账号改投备用/直连。
// 除 expires_at 到期外，也会对 active 代理做连通性测活：一旦测不到延迟就按 fallback 切换。
type ProxyExpiryService struct {
	proxyRepo ProxyRepository
	prober    ProxyExitInfoProber
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewProxyExpiryService(proxyRepo ProxyRepository, interval time.Duration) *ProxyExpiryService {
	return &ProxyExpiryService{proxyRepo: proxyRepo, interval: interval, stopCh: make(chan struct{})}
}

// SetProxyProber enables automatic proxy latency probing + failover on dead proxies.
func (s *ProxyExpiryService) SetProxyProber(prober ProxyExitInfoProber) {
	if s == nil {
		return
	}
	s.prober = prober
}

func (s *ProxyExpiryService) Start() {
	if s == nil || s.proxyRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *ProxyExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.wg.Wait()
}

func (s *ProxyExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	changed, err := s.proxyRepo.SweepExpiredProxies(ctx, time.Now())
	if err != nil {
		log.Printf("[ProxyExpiry] sweep expired proxies failed: %v", err)
	} else if changed > 0 {
		log.Printf("[ProxyExpiry] re-routed %d accounts off expired proxies", changed)
	}

	healthChanged, healthErr := s.probeAndFailoverUnhealthyProxies(ctx)
	if healthErr != nil {
		log.Printf("[ProxyHealth] probe/failover failed: %v", healthErr)
		return
	}
	if healthChanged > 0 {
		log.Printf("[ProxyHealth] re-routed %d accounts off unhealthy proxies", healthChanged)
	}
}

func (s *ProxyExpiryService) probeAndFailoverUnhealthyProxies(ctx context.Context) (int64, error) {
	if s == nil || s.proxyRepo == nil || s.prober == nil {
		return 0, nil
	}

	proxies, err := s.proxyRepo.ListActive(ctx)
	if err != nil {
		return 0, err
	}
	if len(proxies) == 0 {
		return 0, nil
	}

	type probeResult struct {
		proxy Proxy
		err   error
	}

	jobs := make(chan Proxy)
	results := make(chan probeResult, len(proxies))
	workerCount := defaultProxyHealthProbeConcurrency
	if workerCount > len(proxies) {
		workerCount = len(proxies)
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for proxy := range jobs {
				if err := ctx.Err(); err != nil {
					results <- probeResult{proxy: proxy, err: err}
					continue
				}
				probeCtx, cancel := context.WithTimeout(ctx, defaultProxyHealthProbeTimeout)
				_, latencyMs, probeErr := s.prober.ProbeProxy(probeCtx, proxy.URL())
				cancel()
				// "没延迟" = 探测失败或拿不到有效 latency。
				if probeErr == nil && latencyMs <= 0 {
					probeErr = errProxyLatencyMissing
				}
				results <- probeResult{proxy: proxy, err: probeErr}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, proxy := range proxies {
			select {
			case <-ctx.Done():
				return
			case jobs <- proxy:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalChanged int64
	for result := range results {
		if result.err == nil {
			continue
		}
		log.Printf("[ProxyHealth] proxy %d (%s) has no latency: %v", result.proxy.ID, result.proxy.Name, result.err)
		changed, markErr := s.proxyRepo.MarkProxyUnhealthyAndReroute(ctx, result.proxy.ID, time.Now())
		if markErr != nil {
			log.Printf("[ProxyHealth] failover proxy %d failed: %v", result.proxy.ID, markErr)
			continue
		}
		totalChanged += changed
	}
	return totalChanged, nil
}
