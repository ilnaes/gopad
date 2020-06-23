package common

type ResType int

const (
	Ack ResType = iota
	DocRes
	OpsRes
	Error
	Outdated
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
	Seq  int    // last seen seq (always included)
	Doc  Doc    // current document for DocRes
	Ops  [][]Op // ops since last view
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
	Body  string
	View  int
	DocId int64
}

func (d Doc) ApplyOps(op []Op) {

}

func (d Doc) Copy() Doc {
	res := Doc{
		View:  d.View,
		DocId: d.DocId,
		Body:  d.Body,
	}

	return res
}
