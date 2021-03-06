package internal

import (
	"sync"
)

const (
	Ack      = "Ack"
	DocRes   = "DocRes"
	OpsRes   = "OpsRes"
	Error    = "Error"
	Outdated = "Outdated"
)

const (
	Add  = "Add"
	Del  = "Del"
	NOOP = "NOOP"
)

type Request struct {
	IsQuery bool
	DocId   int64
	Uid     string

	Ops  []Op
	View int
	Seq  int

	Num int
}

type Response struct {
	Type string
	View int
	Seq  int // last seen seq (always included)

	Body string // current document for DocRes
	Ops  [][]Op // ops since last view
}

type Op struct {
	Loc  int
	Type string
	Ch   byte
	Seq  int // purely for client-side

	Uid     string
	Session int64
}

type DocMeta struct {
	Doc Doc

	Log         [][]Op         // one update is a collection of ops from one diff
	NextSeq     map[string]int // expected next seq from user
	AppliedSeq  map[string]int // all seqs up to this from user have been applied
	NextDiscard int
	DocId       int64

	mu sync.Mutex // protects individual doc, must hold RLock of server.docs
}

type Doc struct {
	Body  []byte
	View  int
	DocId int64
}

func (d *Doc) ApplyOps(op []Op) {
	d.Body = Apply(d.Body, op)
	d.View++
}
