## tx

A package to commit transactions with functions and eventually rollback.

The idea is that you have a slice of functions (the tasks involved in the transaction), and you want either all of them to be executed, or none.

If one of the functions returns an error, and the next ones depend on the changes of the failed one, how do you recover from that?  
This package eases the recovery from such situations by providing the `Task` struct.  

A `Task` may contain:
 - Func: the main function to be executed.
 - Rollbacks: the additional functions to be called when the Func fails (returns a non-nil error).

You can pass a list of `Task` instances to the `Commit` function, so it will take care of the entire process.

By design, this requires the main function `Func` to have an error in its return types, the position doesn't matter.

You can find examples in `transaction_test.go`.
