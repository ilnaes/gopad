package internal

import "fmt"

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

type Doc struct {
	Body  []byte
	View  int
	DocId int64
}

func (d *Doc) ApplyOps(op []Op) {
	d.Body = Apply(d.Body, op)
	fmt.Printf("DOC: %+v\n%s\n", op, string(d.Body))
	d.View++
}
