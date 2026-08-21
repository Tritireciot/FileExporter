package pgutils

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// LogLevel определяет уровень логирования
type LogLevel int

const (
	ErrorLogLevel LogLevel = iota
	WarnLogLevel
	InfoLogLevel
	DebugLogLevel
)

// BackgroundWorker — воркер с распределённой блокировкой через PostgreSQL.
// Запускает workerFunc только если удалось захватить уникальный лок.
// При потере лока автоматически останавливает полезную нагрузку и пытается перехватить лок снова.
type BackgroundWorker struct {
	// Общий локальный сигнал на польную остановку объекта без возможности перезапуска
	lockingDone chan struct{}

	pool *pgxpool.Pool

	// Мьютекс для защиты разделяемых полей
	mu sync.Mutex

	// Синхронизация полезной нагрузки
	workerDone       chan bool          // сигнал остановки для workerFunc (закрытый = стоп)
	workerDoneClosed bool               // флаг что канал уже закрыт
	workerWait       sync.WaitGroup     // ожидание завершения workerFunc
	workerCtx        context.Context    // контекст для workerFunc
	workerCancel     context.CancelFunc // отмена контекста

	// Функция запуска полезной нагрузки
	// ctx - конекст работы внутренних процессов
	// wg - использовать для ожидания остановки внутренних процессов
	// done - использовать только для проверки на сигнал прерывания процесса
	workerFunc func(ctx context.Context, wg *sync.WaitGroup, done chan bool)
	loggerFunc func(t LogLevel, msg string)

	lockUniqueID string
}

// NewBackgroundWorker создаёт новый экземпляр воркера.
// wfunc: функция полезной нагрузки (обязательна)
// logFunc: опциональная функция логирования (может быть nil)
func NewBackgroundWorker(pool *pgxpool.Pool, uniqueIDString string, wfunc func(ctx context.Context, wg *sync.WaitGroup, done chan bool), logFunc func(t LogLevel, msg string)) *BackgroundWorker {
	if wfunc == nil {
		log.Fatal("BackgroundWorker: workerFunc in null")
	}

	return &BackgroundWorker{
		pool:         pool,
		lockingDone:  make(chan struct{}),
		lockUniqueID: uniqueIDString,
		workerFunc:   wfunc,
		loggerFunc:   logFunc,
	}
}

// -------------------------------------------------------------------
func (w *BackgroundWorker) tryLock() (*PGXSQLLocker, error) {
	var (
		locker *PGXSQLLocker
		err    error
	)
	// Пытаемся создать локер и заблокировать
	if locker, err = CreatePGXSQLLocker(w.pool); err != nil {
		if w.loggerFunc != nil {
			w.loggerFunc(ErrorLogLevel, err.Error())
		}
		return nil, err
	} else {
		if err := locker.TryLock(w.lockUniqueID); err != nil {
			log.Printf("Can't acquire lock: %v", err)
			locker.Close()
			return nil, err
		} else {
			if w.loggerFunc != nil {
				w.loggerFunc(InfoLogLevel, "Lock acquired")
			}
			return locker, nil
		}
	}
}

// -------------------------------------------------------------------
func (w *BackgroundWorker) Start() {
	// Процедура блокировки ресурса
	go func() {
		// Общий цикл
		for {
			var locker *PGXSQLLocker
			var err error

			w.mu.Lock()
			w.workerDone = make(chan bool)
			w.mu.Unlock()
			// Пытаемся блокировать
			for {
				if locker, err = w.tryLock(); err == nil {
					// Если успешно заблокировали - включаем полезную назрузку и выходим из цикла
					go w.runWorker()

					break
				}
				select {
				case <-w.lockingDone:
					return // Прерываем работу в случае завершения работы приложения
				case <-time.After(5 * time.Second):
				}
			}
			//-----------------------------
			// Периодически проверяем что не отлетели
			for {
				if result, err := locker.TestLock(w.lockUniqueID); err != nil {
					if w.loggerFunc != nil {
						w.loggerFunc(ErrorLogLevel, fmt.Sprintf("Locker broken: %v", err))
					}
					break
				} else if result != LockResult_self {
					if w.loggerFunc != nil {
						w.loggerFunc(InfoLogLevel, "Lock lost: held by another instance")
					}
					break // Если отлетели - начинаем бловировать заново
				}
				// Ожидаем до следующей проверки
				select {
				case <-w.lockingDone:
					w.stopWorking()
					locker.Close()
					return // Прерываем работу в случае завершения работы приложения
				case <-time.After(5 * time.Second):
				}
			}
			// Завершаем работу полезной нагрузки
			w.stopWorking()
			locker.Close()
			// Цикл продолжается — пытаемся перехватить блокировку снова
		}
	}()
}

// обёртка для запуска пользовательской функции
func (w *BackgroundWorker) runWorker() {
	w.workerCtx, w.workerCancel = context.WithCancel(context.Background())

	w.workerWait.Add(1)
	defer w.workerWait.Done()

	if w.loggerFunc != nil {
		w.loggerFunc(InfoLogLevel, "Work started")
	}
	// Запуск пользовательской функции
	w.workerFunc(w.workerCtx, &w.workerWait, w.workerDone)

	<-w.workerDone

	if w.workerCancel != nil {
		w.workerCancel()
	}

	if w.loggerFunc != nil {
		w.loggerFunc(InfoLogLevel, "Work finished")
	}
}

// безопасная остановка полезной нагрузки
func (w *BackgroundWorker) stopWorking() {
	w.mu.Lock()
	if w.workerCancel != nil {
		w.workerCancel()
		w.workerCancel = nil
	}
	if w.workerDone != nil && !w.workerDoneClosed {
		close(w.workerDone)
		w.workerDoneClosed = true
	}
	w.mu.Unlock()

	w.workerWait.Wait()
}

// -------------------------------------------------------------------
// Остановка без возможности перезапуска
func (w *BackgroundWorker) Stop() {
	// Сигнализируем об остановке (идемпотентно)
	select {
	case <-w.lockingDone:
		// Уже закрыт — ничего не делаем
	default:
		close(w.lockingDone)
	}

	w.stopWorking()
}
