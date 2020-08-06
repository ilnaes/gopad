package internal

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	s     *Server
	doc   *DocMeta
	uid   string
	conn  *websocket.Conn
	alive bool

	sync.Mutex // protects concurrent conn writes
}

// thread-safe websocket writing
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

	seq := c.doc.AppliedSeq[c.uid]
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
	if m.Seq != seq {
		// incorrect sequence number
		// TODO: figure out error flagging
		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(Response{
			Type: Error,
			Seq:  seq,
		})
	} else if m.View < c.doc.Doc.View-len(c.doc.Log) {
		// view is from too long ago
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
		c.s.cl.Lock()
		// something new
		m.Num = c.s.LastCommit
		c.s.LastCommit++

		c.s.CommitLog = append(c.s.CommitLog, m)
		c.doc.NextSeq[c.uid] = m.Seq + 1
		c.s.cl.Unlock()

		c.doc.mu.Unlock()
		c.s.docs.RUnlock()

		c.write(Response{
			Type: Ack,
			Seq:  seq,
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
