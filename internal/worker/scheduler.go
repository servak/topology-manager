package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Task represents a scheduled task
type Task struct {
	ID          string
	Name        string
	Description string
	Function    func(ctx context.Context) error
	Interval    time.Duration
	Timeout     time.Duration
	LastRun     time.Time
	NextRun     time.Time
	Enabled     bool
	RunCount    int64
	ErrorCount  int64
	LastError   error
}

// TaskStatus represents the status of a task
type TaskStatus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	Interval    string    `json:"interval"`
	LastRun     time.Time `json:"last_run"`
	NextRun     time.Time `json:"next_run"`
	RunCount    int64     `json:"run_count"`
	ErrorCount  int64     `json:"error_count"`
	LastError   string    `json:"last_error,omitempty"`
	IsRunning   bool      `json:"is_running"`
}

// Scheduler manages and executes scheduled tasks
type Scheduler struct {
	tasks       map[string]*Task
	runningTasks map[string]context.CancelFunc
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	logger      *log.Logger
}

// NewScheduler creates a new task scheduler
func NewScheduler(logger *log.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	if logger == nil {
		logger = log.Default()
	}

	return &Scheduler{
		tasks:        make(map[string]*Task),
		runningTasks: make(map[string]context.CancelFunc),
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
	}
}

// AddTask adds a new task to the scheduler
func (s *Scheduler) AddTask(task *Task) error {
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if task.Function == nil {
		return fmt.Errorf("task function cannot be nil")
	}
	if task.Interval <= 0 {
		return fmt.Errorf("task interval must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Set default timeout if not specified
	if task.Timeout <= 0 {
		task.Timeout = 5 * time.Minute
	}

	// Set initial next run time to now for immediate execution
	task.NextRun = time.Now()

	s.tasks[task.ID] = task
	s.logger.Printf("Added task: %s (%s) with interval %s", task.ID, task.Name, task.Interval)

	return nil
}

// RemoveTask removes a task from the scheduler
func (s *Scheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// Cancel running task if any
	if cancel, running := s.runningTasks[taskID]; running {
		cancel()
		delete(s.runningTasks, taskID)
	}

	delete(s.tasks, taskID)
	s.logger.Printf("Removed task: %s (%s)", taskID, task.Name)

	return nil
}

// EnableTask enables a task
func (s *Scheduler) EnableTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	task.Enabled = true
	task.NextRun = time.Now().Add(task.Interval)
	s.logger.Printf("Enabled task: %s (%s)", taskID, task.Name)

	return nil
}

// DisableTask disables a task
func (s *Scheduler) DisableTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	task.Enabled = false
	
	// Cancel running task if any
	if cancel, running := s.runningTasks[taskID]; running {
		cancel()
		delete(s.runningTasks, taskID)
	}

	s.logger.Printf("Disabled task: %s (%s)", taskID, task.Name)

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.logger.Println("Starting task scheduler...")

	s.wg.Add(1)
	go s.run()
}

// Stop stops the scheduler and waits for all tasks to complete
func (s *Scheduler) Stop() {
	s.logger.Println("Stopping task scheduler...")
	s.cancel()
	s.wg.Wait()
	s.logger.Println("Task scheduler stopped")
}

// GetTaskStatus returns the status of all tasks
func (s *Scheduler) GetTaskStatus() []TaskStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var statuses []TaskStatus
	for _, task := range s.tasks {
		status := TaskStatus{
			ID:          task.ID,
			Name:        task.Name,
			Description: task.Description,
			Enabled:     task.Enabled,
			Interval:    task.Interval.String(),
			LastRun:     task.LastRun,
			NextRun:     task.NextRun,
			RunCount:    task.RunCount,
			ErrorCount:  task.ErrorCount,
			IsRunning:   s.isTaskRunning(task.ID),
		}

		if task.LastError != nil {
			status.LastError = task.LastError.Error()
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// RunTaskNow runs a task immediately (outside of its schedule)
func (s *Scheduler) RunTaskNow(taskID string) error {
	s.mu.RLock()
	task, exists := s.tasks[taskID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	if !task.Enabled {
		return fmt.Errorf("task %s is disabled", taskID)
	}

	if s.isTaskRunning(taskID) {
		return fmt.Errorf("task %s is already running", taskID)
	}

	go s.executeTask(task, true)
	return nil
}

// Private methods

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Println("Scheduler context cancelled, stopping...")
			s.cancelAllRunningTasks()
			return
		case <-ticker.C:
			s.checkAndRunTasks()
		}
	}
}

func (s *Scheduler) checkAndRunTasks() {
	now := time.Now()

	s.mu.RLock()
	var tasksToRun []*Task
	for _, task := range s.tasks {
		if task.Enabled && now.After(task.NextRun) && !s.isTaskRunning(task.ID) {
			tasksToRun = append(tasksToRun, task)
		}
	}
	s.mu.RUnlock()

	for _, task := range tasksToRun {
		go s.executeTask(task, false)
	}
}

func (s *Scheduler) executeTask(task *Task, manualRun bool) {
	taskCtx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	// Register running task
	s.mu.Lock()
	s.runningTasks[task.ID] = cancel
	s.mu.Unlock()

	// Unregister when done
	defer func() {
		s.mu.Lock()
		delete(s.runningTasks, task.ID)
		s.mu.Unlock()
	}()

	runType := "scheduled"
	if manualRun {
		runType = "manual"
	}

	s.logger.Printf("Starting %s run of task: %s (%s)", runType, task.ID, task.Name)
	start := time.Now()

	// Execute the task
	err := task.Function(taskCtx)
	duration := time.Since(start)

	// Update task statistics
	s.mu.Lock()
	task.LastRun = start
	task.RunCount++
	if err != nil {
		task.ErrorCount++
		task.LastError = err
		s.logger.Printf("Task %s (%s) failed after %s: %v", task.ID, task.Name, duration, err)
	} else {
		task.LastError = nil
		s.logger.Printf("Task %s (%s) completed successfully in %s", task.ID, task.Name, duration)
	}

	// Schedule next run (only for scheduled runs)
	if !manualRun {
		task.NextRun = start.Add(task.Interval)
	}
	s.mu.Unlock()
}

func (s *Scheduler) isTaskRunning(taskID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, running := s.runningTasks[taskID]
	return running
}

func (s *Scheduler) cancelAllRunningTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Printf("Cancelling %d running tasks...", len(s.runningTasks))
	for taskID, cancel := range s.runningTasks {
		s.logger.Printf("Cancelling task: %s", taskID)
		cancel()
	}
	s.runningTasks = make(map[string]context.CancelFunc)
}

// TaskBuilder helps build tasks with a fluent interface
type TaskBuilder struct {
	task *Task
}

// NewTaskBuilder creates a new task builder
func NewTaskBuilder(id, name string) *TaskBuilder {
	return &TaskBuilder{
		task: &Task{
			ID:      id,
			Name:    name,
			Enabled: true,
			Timeout: 5 * time.Minute,
		},
	}
}

// Description sets the task description
func (tb *TaskBuilder) Description(desc string) *TaskBuilder {
	tb.task.Description = desc
	return tb
}

// Interval sets the task interval
func (tb *TaskBuilder) Interval(interval time.Duration) *TaskBuilder {
	tb.task.Interval = interval
	return tb
}

// Timeout sets the task timeout
func (tb *TaskBuilder) Timeout(timeout time.Duration) *TaskBuilder {
	tb.task.Timeout = timeout
	return tb
}

// Function sets the task function
func (tb *TaskBuilder) Function(fn func(ctx context.Context) error) *TaskBuilder {
	tb.task.Function = fn
	return tb
}

// Enabled sets whether the task is enabled
func (tb *TaskBuilder) Enabled(enabled bool) *TaskBuilder {
	tb.task.Enabled = enabled
	return tb
}

// Build returns the built task
func (tb *TaskBuilder) Build() *Task {
	return tb.task
}