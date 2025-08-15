package concurrency

// ConcurrencyLimiterRunConfig allows you to configure the behavior of the ConcurrencyLimiter.Run() function
type ConcurrencyLimiterRunConfig struct {
	onComplete func()
}

type ConcurrencyLimiterRunOption func(cfg ConcurrencyLimiterRunConfig) ConcurrencyLimiterRunConfig

// WithOnCompleteCallback - This option will allow you to specify a callback function that will be called after the goroutine exits
func WithOnCompleteCallback(onComplete func()) ConcurrencyLimiterRunOption {
	return func(cfg ConcurrencyLimiterRunConfig) ConcurrencyLimiterRunConfig {
		cfg.onComplete = onComplete
		return cfg
	}
}

// ConcurrencyLimiter will allow you to manage a maximum number of concurrently executing goroutines
type ConcurrencyLimiter struct {
	maxConcurrency       int
	currentConcurrency   int
	orchestrationChannel chan struct{}
}

func NewConcurrencyLimiter(maxConcurrency int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		maxConcurrency:       maxConcurrency,
		orchestrationChannel: make(chan struct{}, maxConcurrency),
	}
}

// Run will schedule the current function as a go routine if concurrency is available. If no concurrency is available, this function will block
// until other scheduled goroutines exit. If concurrency is available, the routine will be started and the Run() will immediately return.
func (rp *ConcurrencyLimiter) Run(routine func(), opts ...ConcurrencyLimiterRunOption) {
	// This will block until currency is available
	rp.incrementConcurrency()

	cfg := ConcurrencyLimiterRunConfig{}
	for _, opt := range opts {
		cfg = opt(cfg)
	}

	go func() {
		routine()
		rp.decrementConcurrency()

		// If on complete callback is set, run it after goroutine exits
		if cfg.onComplete != nil {
			cfg.onComplete()
		}
	}()

}

func (rp *ConcurrencyLimiter) incrementConcurrency() {
	rp.orchestrationChannel <- struct{}{}
}

func (rp *ConcurrencyLimiter) decrementConcurrency() {
	<-rp.orchestrationChannel
}
