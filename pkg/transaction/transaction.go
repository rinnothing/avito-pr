package transaction

import "context"

func DoAtomically(func(context.Context) error) error {
	panic("not implemented")
}
