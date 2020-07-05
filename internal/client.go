package internal

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	s     *Server
	doc   *DocMeta
	uid   int64
	conn  *websocket.Conn
	alive bool

	sync.Mutex // protects concurrent conn writes
}

func (c *Client) write(res Response) {
	c.Lock()
	err := c.conn.WriteJSON(res)
	c.Unlock()
	if err != nil {
		c.alive = false
	}
}

// handles document queries
func (c *Client) handleQuery(view int) {
	c.s.docs.RLock()
	c.doc.mu.Lock()

	seq := c.doc.NextSeq[c.uid]
	if view > c.doc.Doc.View {
		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(Response{
			Type: Error,
			Seq:  seq,
		})
	} else if view < c.doc.Doc.View-len(c.doc.Log) {
		// view is from before the beginning of log
		res := Response{
			Type: DocRes,
			Body: string(c.doc.Doc.Body),
			View: c.doc.Doc.View,
			Seq:  seq,
		}

		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(res)
	} else {
		// send log ops
		res := make([][]Op, c.doc.Doc.View-view)
		copy(res, c.doc.Log[len(c.doc.Log)-(c.doc.Doc.View-view):])

		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(Response{
			Type: OpsRes,
			View: view,
			Ops:  res,
			Seq:  seq,
		})
	}
}

// does basic checking and adds an Op slice to commit log
func (c *Client) handleOps(m Request) {
	// TODO: welldef check ops (check docId, uid, seq ordering)

	if len(m.Ops) == 0 {
		return
	}

	c.s.docs.RLock()
	c.doc.mu.Lock()

	seq := c.doc.NextSeq[c.uid]
	if m.Ops[0][0].Seq > seq+1 {
		// too high sequence number
		// TODO: figure out error flagging
		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(Response{
			Type: Error,
			Seq:  seq,
		})
	} else if m.View < c.doc.Doc.View-len(c.doc.Log) {
		// view is from too far ago
		res := Response{
			Type: DocRes,
			Body: string(c.doc.Doc.Body),
			View: c.doc.Doc.View,
			Seq:  seq,
		}

		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(res)
	} else {
		N := len(m.Ops)

		lastSeq := m.Ops[N-1][0].Seq

		c.s.cl.Lock()
		if lastSeq >= c.doc.NextSeq[c.uid] {
			// something new
			c.s.CommitLog = append(c.s.CommitLog, m)
			c.doc.NextSeq[c.uid] = lastSeq + 1
		}
		c.s.cl.Unlock()

		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(Response{
			Type: Ack,
			Seq:  lastSeq,
		})
	}
}

func (c *Client) interact() {
	for c.alive {
		var m Request
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
