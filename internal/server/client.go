package server

import (
	"log"
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

	sync.Mutex // protects concurrent conn writes
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
	seq := c.doc.NextSeq[c.uid]
	if view > c.doc.Doc.View {
		c.s.Unlock()

		c.write(co.Response{
			Type: co.Error,
			Seq:  seq,
		})
	} else if view < c.doc.Doc.View-len(c.doc.Log) {
		// view is from before the beginning of log
		res := co.Response{
			Type: co.DocRes,
			Body: string(c.doc.Doc.Body),
			View: c.doc.Doc.View,
			Seq:  seq,
		}
		c.s.Unlock()

		c.write(res)
	} else {
		// send log ops
		l := len(c.doc.Log) - c.doc.Doc.View + view
		res := make([][]co.Op, l)
		copy(res, c.doc.Log[len(c.doc.Log)-(c.doc.Doc.View-view):])

		c.s.Unlock()

		c.write(co.Response{
			Type: co.OpsRes,
			Ops:  res,
			Seq:  seq,
		})
	}
}

// does basic checking and adds an Op slice to commit log
func (c *Client) handleOps(m co.Request) {
	// TODO: welldef check ops (check docId, uid, seq ordering)

	if len(m.Ops) == 0 {
		return
	}

	c.s.Lock()
	seq := c.doc.NextSeq[c.uid]
	if m.Ops[0][0].Seq > seq+1 {
		// too high sequence number
		// TODO: figure out error flagging
		c.s.Unlock()

		c.write(co.Response{
			Type: co.Error,
			Seq:  seq,
		})
	} else if m.View < c.doc.Doc.View-len(c.doc.Log) {
		// view is from too far ago
		res := co.Response{
			Type: co.DocRes,
			Body: string(c.doc.Doc.Body),
			View: c.doc.Doc.View,
			Seq:  seq,
		}
		c.s.Unlock()

		c.write(res)
	} else {
		N := len(m.Ops)

		lastSeq := m.Ops[N-1][0].Seq

		if lastSeq >= c.doc.NextSeq[c.uid] {
			// something new
			c.s.CommitLog = append(c.s.CommitLog, m)
			c.doc.NextSeq[c.uid] = lastSeq + 1
		}
		c.s.Unlock()

		c.write(co.Response{
			Type: co.Ack,
			Seq:  lastSeq,
		})
	}
}

func (c *Client) interact() {
	for c.alive {
		var m co.Request
		if err := c.conn.ReadJSON(&m); err != nil {
			log.Println(err)
			break
		}

		if m.IsQuery {
			go c.handleQuery(m.View)
		} else if m.Ops != nil {
			go c.handleOps(m)
		}
	}
}
