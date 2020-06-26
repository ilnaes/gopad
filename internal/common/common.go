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
	Ops     [][]Op
	View    int
	DocId   int64
	Uid     int64
}

type Response struct {
	Type ResType
	Body string // current document for DocRes
	View int
	Ops  [][]Op // ops since last view
	Seq  int    // last seen seq (always included)
}

type Op struct {
	Loc int
	Ch  byte
	Add bool

	DocId int64
	Uid   int64
	Seq   int
}

type Doc struct {
	Body  []byte
	View  int
	DocId int64
}

func (d Doc) ApplyOps(op []Op) {
	d.Body = Apply(d.Body, op)
}
