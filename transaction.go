package tx

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrMustReturnError   = errors.New("the first function of a transaction must return at least one error")
	ErrTransactionFailed = errors.New("transaction failed")
)

// Task represents the tasks of a transaction.
//
// Each task is composed of a function that represents the task to be
// completed, and an optional list of functions to be executed if an error
// is returned from the first function after its execution.
type Task struct {
	Func      interface{}
	Rollbacks []interface{}
}

// Commit takes care of executing the transaction's tasks.
//
// The function reads all the tasks one by one and executes
// their first function [Task].Func.
// If the first function returns an error, all the additional
// functions defined in FuncList will be called, and so will be
// the additional functions of previous tasks, as a way to rollback
// any change committed until that moment.
func Commit(tasks ...Task) error {
	var funcsToExecute []reflect.Value

	// We must loop on tasks twice to make sure
	// that each first-function has an error return type.
	for _, task := range tasks {
		funcType := reflect.TypeOf(task.Func)

		if _, hasErr := hasErrorReturnType(&funcType); !hasErr {
			return ErrMustReturnError
		}
	}

	ok := true
	var taskError error

	for i, task := range tasks {
		funcType := reflect.TypeOf(task.Func)
		funcValue := reflect.ValueOf(task.Func)

		// Find the index of the error return value.
		// Throw an error if the function doesn't have an error
		// in its return types.
		errPos, returnsError := hasErrorReturnType(&funcType)
		if !returnsError {
			break
		}

		// Execute the function
		results := funcValue.Call(nil)
		if !results[errPos].IsNil() {
			taskError = (results[errPos].Interface()).(error)

			// If the function returns an error, append all the additional functions to the funcsToExecute slice
			for k, t := range tasks {
				if k > i {
					break
				}

				for _, f := range t.Rollbacks {
					funcsToExecute = append(funcsToExecute, reflect.ValueOf(f))
				}
			}
			ok = false

			break
		}
	}

	// Execute all the functions in the funcsToExecute slice
	// in reverse order.
	for i := len(funcsToExecute) - 1; i >= 0; i-- {
		(funcsToExecute[i]).Call(nil)
	}

	if !ok {
		return errors.Join(ErrTransactionFailed, fmt.Errorf("task returned an error: %v", taskError))
	}

	return nil
}

func hasErrorReturnType(funcType *reflect.Type) (int, bool) {
	for i := 0; i < (*funcType).NumOut(); i++ {
		if (*funcType).Out(i) == reflect.TypeOf((*error)(nil)).Elem() {
			return i, true
		}
	}

	return -1, false
}
