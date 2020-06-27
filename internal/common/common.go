package common

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
	DocId int64
	Uid   int64
	Seq   int

	Loc int
	Add bool
	Ch  byte
}

type Doc struct {
	Body  []byte
	View  int
	DocId int64
}

func (d Doc) ApplyOps(op []Op) {
	d.Body = Apply(d.Body, op)
}
