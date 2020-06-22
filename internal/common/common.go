package common

type Handshake struct {
	Doc    int64
	Create bool
}

type Op struct {
	Loc int
	Ch  byte
	Add bool

	Doc int
}

type Doc struct {
	Body []byte
	View int64
	Id   int64
}
