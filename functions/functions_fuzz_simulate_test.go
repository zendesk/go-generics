package functions

// This can be removed once the bug in fuzz testing has been remediated and we can prove it. This file is useful for
// simulating fuzz tests while avoiding the fuzz bug.

//	func TestFuzzGoMapToManyManualFuzzing(t *testing.T) {
//		rand.Seed(time.Now().UnixNano())
//		times := 100
//
//		runner := threading.NewTaskRunner(runtime.NumCPU())
//		bar := pb.New(times).Start()
//
//		var wg sync.WaitGroup
//		wg.Add(times)
//		for i := 0; i < times; i++ {
//			runner.Schedule(func() {
//				start := time.Now()
//				defer func() {
//					if time.Since(start) > time.Minute {
//						t.Fatal("timeout")
//					}
//					wg.Done()
//				}()
//
//				t.Run(strconv.Itoa(i), func(t *testing.T) {
//					num := randomNumberBetween(0, 30000)
//
//					foos := MakeFoos(num)
//					barsPerFoo := num
//					if num > 100 {
//						barsPerFoo = 100
//					}
//
//					// remove any ones with duplicate order value b/c that's how we're ordering below.
//					foos = DedupeByHash(foos, hashByOrder)
//
//					// execute
//					var bars []*Bar
//					if num%2 == 0 {
//						bars = GoMapToMany(foos, toManyBars(barsPerFoo))
//					} else {
//						bars = GoMapToMany(foos, toManyBars(barsPerFoo), ConcurrencyLimitOption(num/3+1))
//					}
//
//					// Validate data was mutated properly by provided function
//					var expectedBars []*Bar
//					for _, foo := range foos {
//						expectedBars = append(expectedBars, toManyBars(barsPerFoo)(foo)...)
//					}
//
//					sort.Slice(expectedBars, func(i, j int) bool {
//						return expectedBars[i].Order > expectedBars[j].Order
//					})
//
//					sort.Slice(bars, func(i, j int) bool {
//						return bars[i].Order > bars[j].Order
//					})
//
//					test.CheckEqual(len(bars), "Bar Length is not equal expected bar length", len(expectedBars), t)
//					test.CheckEqual(bars, "Bars", expectedBars, t)
//					bar.Increment()
//				})
//			})
//		}
//		wg.Wait()
//		bar.Finish()
//	}
//func TestFuzzGoMapWithErrsRateLimitTestManualFuzzing(t *testing.T) {
//	rand.Seed(time.Now().UnixNano())
//	times := 200
//
//	runner := threading.NewTaskRunner(runtime.NumCPU())
//	bar := pb.New(times).Start()
//
//	var wg sync.WaitGroup
//	wg.Add(times)
//	for i := 0; i < times; i++ {
//		runner.Schedule(func() {
//			started := time.Now()
//			defer func() {
//				if time.Since(started) > time.Minute {
//					t.Fatal("timeout")
//				}
//				wg.Done()
//			}()
//
//			t.Run(strconv.Itoa(i), func(t *testing.T) {
//				sliceSize := randomNumber(maxSliceSizeLengthRateLimit)
//				// we want to ensure rate < sliceSize otherwise no throttling will occur and we cannot estimate expectedDuration. Also rate cannot be 0
//				rate := randomNumberBetween(minRatePerInterval, (sliceSize+1)/5+1)
//				durationNanoseconds := randomDurationBetween(time.Millisecond, time.Second).Nanoseconds()
//
//				// If rate is very low, reset rate to ensure we don't run TOO long (max 50 seconds with this change)
//				if sliceSize != 0 && sliceSize/rate > 50 {
//					rate = sliceSize / 10
//				}
//
//				ratePerTime := Max(rate, -rate)
//				duration := time.Duration(durationNanoseconds)
//				foos := MakeFoos(sliceSize)
//
//				// estimate expected execution time given rate limit
//				concurrency := randomNumberBetween(1, 20) // Concurrency doesn't matter, rate is limited across goroutines
//
//				var expectedDurationMillis float64
//				// excluding the first batch, we can assume rate-limiting for all subsequent batches at the per-time interval. First batch starts immediately
//				if ratePerTime > 0 {
//					expectedDurationMillis = float64(len(foos)-ratePerTime)/(float64(ratePerTime)/float64(duration.Milliseconds())) - 1
//				} else {
//					// if no rate limit is specified, expect this to be very fast
//					expectedDurationMillis = 0
//				}
//
//				// execute
//				start := time.Now().UnixMilli()
//				_, _ = GoMapWithErrs(foos, toBarWithErr, ConcurrencyLimitOption(concurrency), RateLimitOption(ratePerTime, duration))
//				finish := time.Now().UnixMilli()
//
//				totalTime := float64(finish - start)
//
//				// If no rate limiting was happening, the actual process time would be nanoseconds long
//				// and far below minProcessTime. Min process time is best-case scenario
//				// We cannot reasonably estimate max process time because we're on a system that is loaded by concurrent Fuzz tests
//				// and CPU wait is a real thing. What we _do_ know, is that the test should not finish before minProcessTime elapses
//				minProcessTime := expectedDurationMillis
//
//				if minProcessTime <= totalTime {
//					t.Logf("SUCCESS: Process took %f millis. Expected at least %f", totalTime, minProcessTime)
//				} else {
//					t.Fatalf("FAILURE: Process took %f millis. Expected at least %f", totalTime, minProcessTime)
//				}
//				bar.Increment()
//			})
//		})
//	}
//	wg.Wait()
//	bar.Finish()
//}

//
//func TestGoMapWithErrsRateLimitTest(t *testing.T) {
//	ratePerTime := 18
//	durationNanoseconds := int64(343012204)
//	num := 756
//
//	ratePerTime = Max(ratePerTime, -ratePerTime)
//	duration := time.Duration(durationNanoseconds)
//	foos := MakeFoos(num)
//
//	// estimate expected execution time given rate limit
//	concurrency := randomNumberBetween(1, 20) // Concurrency doesn't matter, rate is limited across goroutines
//
//	var expectedDurationMillis float64
//	// excluding the first batch, we can assume rate-limiting for all subsequent batches at the per-time interval. First batch starts immediately
//	if ratePerTime > 0 {
//		expectedDurationMillis = float64(len(foos)-ratePerTime)/(float64(ratePerTime)/float64(duration.Milliseconds())) - 1
//	} else {
//		// if no rate limit is specified, expect this to be very fast
//		expectedDurationMillis = 0
//	}
//
//	// execute
//	start := time.Now().UnixMilli()
//	_, _ = GoMapWithErrs(foos, toBarWithErr, ConcurrencyLimitOption(concurrency), RateLimitOption(ratePerTime, duration))
//	finish := time.Now().UnixMilli()
//
//	totalTime := float64(finish - start)
//
//	// If no rate limiting was happening, the actual process time would be nanoseconds long
//	// and far below minProcessTime. Min process time is best-case scenario
//	// We cannot reasonably estimate max process time because we're on a system that is loaded by concurrent Fuzz tests
//	// and CPU wait is a real thing. What we _do_ know, is that the test should not finish before minProcessTime elapses
//	minProcessTime := expectedDurationMillis
//
//	if minProcessTime <= totalTime {
//		t.Logf("SUCCESS: Process took %f millis. Expected at least %f", totalTime, minProcessTime)
//	} else {
//		t.Fatalf("FAILURE: Process took %f millis. Expected at least %f", totalTime, minProcessTime)
//	}
//}
