//go:build test
// +build test

package concurrency

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

// Mock implementations for testing
type mockLock struct {
	name string
}

func (m *mockLock) Release(ctx context.Context) error {
	return nil
}

func (m *mockLock) Refresh(ctx context.Context) error {
	return nil
}

type mockLockBackend struct {
	obtainLockFunc func(ctx context.Context, name string, ttl time.Duration) (Lock, error)
}

func (m *mockLockBackend) ObtainLock(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
	if m.obtainLockFunc != nil {
		return m.obtainLockFunc(ctx, name, ttl)
	}
	return &mockLock{name: name}, nil
}

type mockLockBackendWithCounter struct {
	counter        *int // Change to pointer to ensure persistence
	obtainLockFunc func(counter *int, ctx context.Context, name string, ttl time.Duration) (Lock, error)
}

func (m *mockLockBackendWithCounter) ObtainLock(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
	if m.obtainLockFunc != nil {
		return m.obtainLockFunc(m.counter, ctx, name, ttl)
	}
	return &mockLock{name: name}, nil
}

func TestLockManager_retryAcquireWithBackoff(t *testing.T) {
	type fields struct {
		backend LockBackend
	}
	type args struct {
		ctx     context.Context
		key     string
		lockTTL time.Duration
		timeout time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Lock
		wantErr bool
	}{
		{
			name: "successful acquisition on first retry",
			fields: fields{
				backend: &mockLockBackend{
					obtainLockFunc: func(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
						return &mockLock{name: name}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				key:     "test-key",
				lockTTL: time.Minute,
				timeout: time.Second,
			},
			want:    &mockLock{name: "test-key"},
			wantErr: false,
		},
		{
			name: "timeout waiting for lock",
			fields: fields{
				backend: &mockLockBackend{
					obtainLockFunc: func(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
						return nil, ErrorLockNotAcquired
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				key:     "test-key",
				lockTTL: time.Minute,
				timeout: 100 * time.Millisecond, // Short timeout to trigger timeout
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "context cancellation",
			fields: fields{
				backend: &mockLockBackend{
					obtainLockFunc: func(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
						return nil, ErrorLockNotAcquired
					},
				},
			},
			args: args{
				ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
				key:     "test-key",
				lockTTL: time.Minute,
				timeout: time.Second,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "backend error during acquisition",
			fields: fields{
				backend: &mockLockBackend{
					obtainLockFunc: func(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
						return nil, errors.New("backend error")
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				key:     "test-key",
				lockTTL: time.Minute,
				timeout: time.Second,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "successful acquisition after multiple retries",
			fields: fields{
				backend: &mockLockBackendWithCounter{
					counter: new(int),
					obtainLockFunc: func(counter *int, ctx context.Context, name string, ttl time.Duration) (Lock, error) {
						*counter++
						if *counter < 3 {
							return nil, ErrorLockNotAcquired
						}
						return &mockLock{name: name}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				key:     "test-key",
				lockTTL: time.Minute,
				timeout: time.Second * 10,
			},
			want:    &mockLock{name: "test-key"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := &LockManager{
				backend: tt.fields.backend,
			}
			got, err := lm.retryAcquireWithBackoff(tt.args.ctx, tt.args.key, tt.args.lockTTL, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("LockManager.retryAcquireWithBackoff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LockManager.retryAcquireWithBackoff() = %v, want %v", got, tt.want)
			}
		})
	}
}
