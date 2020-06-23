package server

import (
	"sync"

	"github.com/gorilla/websocket"
	co "github.com/ilnaes/gopad/internal/common"
)

type Client struct {
	s     *Server
	doc   *DocMeta
	uid   int64
	conn  *websocket.Conn
	alive bool

	sync.Mutex
}

func (c *Client) write(res co.Response) {
	c.Lock()
	err := c.conn.WriteJSON(res)
	c.Unlock()
	if err != nil {
		c.alive = false
	}
}

// handles document queries
func (c *Client) handleQuery(view int) {
	c.s.Lock()
	if view > c.doc.Doc.View {
		c.s.Unlock()
		c.write(co.Response{
			Type: co.Error,
		})
	} else if view < c.doc.Doc.View-len(c.doc.Log) {
		// TODO: return whole document
	} else {
		res := append([]co.Op{}, c.doc.Log[len(c.doc.Log)-(c.doc.Doc.View-view):]...)
		c.s.Unlock()
		c.write(co.Response{
			Type: co.OpsRes,
			Ops:  res,
		})
	}
}

// does basic checking and adds an Op slice to commit log
func (c *Client) handleOps(m co.Request) {
	// TODO: welldef check ops (check docId, uid, seq ordering)

	c.s.Lock()
	if m.Ops[0].Seq > c.doc.SeenSeqs[c.uid]+1 {
		// too high sequence number
		// TODO: figure out error flagging
		c.s.Unlock()
		c.write(co.Response{
			Type: co.Error,
		})
		return
	}

	N := len(m.Ops)
	if m.Ops[N-1].Seq > c.doc.SeenSeqs[c.uid] {
		// something new
		c.s.CommitLog = append(c.s.CommitLog, m)
		c.doc.SeenSeqs[c.uid] = m.Ops[N-1].Seq
	}
	c.s.Unlock()

	c.write(co.Response{
		Type: co.Ack,
		Seq:  m.Ops[len(m.Ops)-1].Seq,
	})
}

func (c *Client) interact() {
	for c.alive {
		var m co.Request
		c.Lock()
		err := c.conn.ReadJSON(&m)
		c.Unlock()
		if err != nil {
			break
		}

		if m.IsQuery {
			go c.handleQuery(m.View)
		} else if m.Ops != nil {
			go c.handleOps(m)
		}
	}
}
