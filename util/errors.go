package util

type LbErr struct {
    ParentErr error
    Code string
    Desc string
}

func (e *LbErr) Error() string {
    return e.Code + " " + e.Desc
}

func (e *LbErr) Parent() error {
    return e.ParentErr
}

func NewLbErr(parent error, code string, desp string) *LbErr {
    return &LbErr{parent, code, desp}
}
