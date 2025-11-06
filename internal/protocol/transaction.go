package protocol

type Transaction struct {
	Operations []*Operation
}

// This is not exhaustive, it is to be used in conjunction with parseTransaction()
// A Transaction can contain only write Operations (SET and DELETE)
// A Transaction can only have 8 operations (this is arbitrary)
func (tx *Transaction) IsValid() bool {
	for _, op := range tx.Operations {
		if op.Directive != SetDirective || op.Directive != DeleteDirective {
			return false
		}
	}

	if len(tx.Operations) > 8 || len(tx.Operations) == 0 {
		return false
	}

	return true
}
