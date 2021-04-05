package tx

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTask is a mock implementation of the Task struct.
type MockTask struct {
	mock.Mock
}

func (m *MockTask) Run() error {
	args := m.Called()
	return args.Error(0)
}

func TestCommit(t *testing.T) {
	t.Run("no tasks", func(t *testing.T) {
		err := Commit()
		assert.NoError(t, err)
	})

	t.Run("tasks with no errors", func(t *testing.T) {
		task1 := Task{
			Func: func() error { return nil },
		}
		task2 := Task{
			Func: func() error { return nil },
		}

		err := Commit(task1, task2)
		assert.NoError(t, err)
	})

	t.Run("task with error", func(t *testing.T) {
		task1 := Task{
			Func: func() error { return errors.New("error") },
		}
		task2 := Task{
			Func: func() error { return nil },
		}

		err := Commit(task1, task2)
		assert.Error(t, err)
		assert.EqualError(t, err, "transaction failed\ntask returned an error: error")
	})

	t.Run("task with error and rollback", func(t *testing.T) {
		var rollback1Called bool
		var rollback2Called bool

		task1 := Task{
			Func: func() error { return errors.New("error") },
			Rollbacks: []interface{}{
				func() { rollback1Called = true },
			},
		}
		task2 := Task{
			Func: func() error { return nil },
			Rollbacks: []interface{}{
				func() { rollback2Called = true },
			},
		}

		err := Commit(task1, task2)
		assert.Error(t, err)
		assert.EqualError(t, err, "transaction failed\ntask returned an error: error")
		assert.True(t, rollback1Called)
		assert.False(t, rollback2Called)
	})

	t.Run("task with no error and rollback", func(t *testing.T) {
		var rollback1Called bool
		var rollback2Called bool

		task1 := Task{
			Func: func() error { return nil },
			Rollbacks: []interface{}{
				func() { rollback1Called = true },
			},
		}
		task2 := Task{
			Func: func() error { return nil },
			Rollbacks: []interface{}{
				func() { rollback2Called = true },
			},
		}

		err := Commit(task1, task2)
		assert.NoError(t, err)
		assert.False(t, rollback1Called)
		assert.False(t, rollback2Called)
	})

	t.Run("task with no error and rollback with error", func(t *testing.T) {
		var rollback1Called bool
		var rollback2Called bool

		task1 := Task{
			Func: func() error { return nil },
			Rollbacks: []interface{}{
				func() { rollback1Called = true },
				func() error { return errors.New("error") },
			},
		}
		task2 := Task{
			Func: func() error { return nil },
			Rollbacks: []interface{}{
				func() { rollback2Called = true },
			},
		}

		err := Commit(task1, task2)
		assert.NoError(t, err)
		assert.False(t, rollback1Called)
		assert.False(t, rollback2Called)
	})
}

func TestHasErrorReturnType(t *testing.T) {
	// Test the case where the function has an error return type.
	funcType := reflect.TypeOf(func() error { return nil })
	_, hasError := hasErrorReturnType(&funcType)
	assert.True(t, hasError)

	// Test the case where the function does not have an error return type.
	funcType = reflect.TypeOf(func() {})
	_, hasError = hasErrorReturnType(&funcType)
	assert.False(t, hasError)
}
