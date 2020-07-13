package internal

import (
	"sync"
)

type ResType string

const (
	Ack      = "Ack"
	DocRes   = "DocRes"
	OpsRes   = "OpsRes"
	Error    = "Error"
	Outdated = "Outdated"
)

type Request struct {
	IsQuery bool
	View    int
	DocId   int64
	Uid     int64

	Ops [][]Op
	Num int
}

type Response struct {
	Type ResType
	View int
	Seq  int // last seen seq (always included)

	Body string // current document for DocRes
	Ops  [][]Op // ops since last view
}

type Op struct {
	Loc int
	Add bool
	Ch  byte

	Seq  int
	Uid  int64
	View int
}

type DocMeta struct {
	Doc Doc

	Log         [][]Op        // one update is a collection of ops from one diff
	NextSeq     map[int64]int // expected next seq from user
	AppliedSeqs map[int64]int // all seqs up to this from user have been applied
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
