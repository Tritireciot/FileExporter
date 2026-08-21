package pgutils

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := NewPool(&DatabaseConfig{
		Host:     "10.0.2.212",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBname:   "studio_pinchuk",
		Schema:   "config",
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	return pool
}

// -------------------------------------------------------------------
// Тест 1: 2 воркера с разными идентификаторами — успешный захват и одновременная работа
func TestUniqueWorker_TwoDifferentWorkers(t *testing.T) {
	var w1Count int64
	var w2Count int64

	pgPool1 := newTestPool(t)
	pgPool2 := newTestPool(t)

	w1 := NewBackgroundWorker(pgPool1, "different-id-a11", func(ctx context.Context, wg *sync.WaitGroup, done chan bool) {
		for {
			atomic.AddInt64(&w1Count, 1)
			log.Printf("1: tick: %d", w1Count)
			select {
			case <-ctx.Done():
				log.Printf("w1: ctx done")
				return
			case <-done:
				log.Printf("w1: ch done")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}, nil)
	w2 := NewBackgroundWorker(pgPool2, "different-id-b22", func(ctx context.Context, wg *sync.WaitGroup, done chan bool) {
		for {
			atomic.AddInt64(&w2Count, 1)
			log.Printf("2: tick: %d", w2Count)
			select {
			case <-ctx.Done():
				log.Printf("w2: ctx done")
				return
			case <-done:
				log.Printf("w2: ch done")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}, nil)

	w1.Start()
	w2.Start()

	time.Sleep(1 * time.Second)

	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			w1.Stop()
			wg.Done()
		}()
		go func() {
			w2.Stop()
			wg.Done()
		}()
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
		break
	}

	final1 := atomic.LoadInt64(&w1Count)
	final2 := atomic.LoadInt64(&w2Count)
	log.Printf("[PASS] Разные ID: w1=%d, w2=%d\n", final1, final2)
	assert.True(t, final1 > 0)
	assert.True(t, final2 > 0)

	time.Sleep(500 * time.Millisecond)
}

// -------------------------------------------------------------------
// Тест 2: 2 воркера с одинаковыми идентификаторами — успешный захват одного, 2й не работает
func TestUniqueWorker_SameID(t *testing.T) {
	var w1Count int64
	var w2Count int64

	pgPool1 := newTestPool(t)
	pgPool2 := newTestPool(t)

	w1 := NewBackgroundWorker(pgPool1, "same-id-123", func(ctx context.Context, wg *sync.WaitGroup, done chan bool) {
		for {
			atomic.AddInt64(&w1Count, 1)
			log.Printf("w1: tick: %d", w1Count)
			select {
			case <-ctx.Done():
				log.Printf("w1: ctx done")
				return
			case <-done:
				log.Printf("w1: ch done")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}, nil)
	w2 := NewBackgroundWorker(pgPool2, "same-id-123", func(ctx context.Context, wg *sync.WaitGroup, done chan bool) {
		for {
			atomic.AddInt64(&w2Count, 1)
			log.Printf("w2: tick: %d", w2Count)
			select {
			case <-ctx.Done():
				log.Printf("w2: ctx done")
				return
			case <-done:
				log.Printf("w2: ch done")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}, nil)

	w1.Start()
	w2.Start()

	time.Sleep(1 * time.Second)

	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			w1.Stop()
			wg.Done()
		}()
		go func() {
			w2.Stop()
			wg.Done()
		}()
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
		break
	}

	final1 := atomic.LoadInt64(&w1Count)
	final2 := atomic.LoadInt64(&w2Count)
	log.Printf("[PASS] Одинаковый ID: w1=%d, w2=%d\n", final1, final2)
	assert.True(t, final1 > 0, "w1 should work")
	assert.True(t, final2 == 0, "w2 should NOT work (cannot acquire lock)")

	time.Sleep(500 * time.Millisecond)
}

// -------------------------------------------------------------------
// Тест 3: 2 воркера с одинаковыми идентификаторами — первый работает, второй ждёт,
// после остановки первого — второй захватывает лок
func TestUniqueWorker_SameID_Handoff(t *testing.T) {
	var w1Count int64
	var w2Count int64

	pgPool1 := newTestPool(t)
	pgPool2 := newTestPool(t)

	w1 := NewBackgroundWorker(pgPool1, "handoff-id-456", func(ctx context.Context, wg *sync.WaitGroup, done chan bool) {
		for {
			atomic.AddInt64(&w1Count, 1)
			log.Printf("w1: tick: %d", w1Count)
			select {
			case <-ctx.Done():
				log.Printf("w1: ctx done")
				return
			case <-done:
				log.Printf("w1: ch done")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}, nil)
	w2 := NewBackgroundWorker(pgPool2, "handoff-id-456", func(ctx context.Context, wg *sync.WaitGroup, done chan bool) {
		for {
			atomic.AddInt64(&w2Count, 1)
			log.Printf("w2: tick: %d", w2Count)
			select {
			case <-ctx.Done():
				log.Printf("w2: ctx done")
				return
			case <-done:
				log.Printf("w2: ch done")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}, nil)

	w1.Start()
	w2.Start()

	time.Sleep(1 * time.Second)

	log.Printf("=== Stopping w1 ===")
	w1.Stop()

	log.Printf("=== Waiting for w2 to acquire lock ===")
	for i := 0; i < 15; i++ {
		time.Sleep(500 * time.Millisecond)
		if atomic.LoadInt64(&w2Count) > 0 {
			log.Printf("w2 started working after %d ms", (i+1)*500)
			break
		}
	}

	time.Sleep(500 * time.Millisecond)

	log.Printf("=== Stopping w2 ===")
	done := make(chan bool)
	go func() {
		w2.Stop()
		done <- true
	}()

	select {
	case <-done:
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
		break
	}

	final1 := atomic.LoadInt64(&w1Count)
	final2 := atomic.LoadInt64(&w2Count)
	log.Printf("[PASS] Handoff: w1=%d, w2=%d\n", final1, final2)
	assert.True(t, final1 > 0, "w1 should work")
	assert.True(t, final2 > 0, "w2 should start working after w1 stops")

	time.Sleep(500 * time.Millisecond)
}
