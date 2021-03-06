package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/joho/godotenv"
)

const (
	UpdateInterval = 250 * time.Millisecond
	SnapMult       = 100
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	Docs       map[int64]*DocMeta
	CommitLog  []Request // Paxos stand-in for now
	LastCommit int       // last req to have been saved to commit log

	cl   sync.Mutex   // protects CommitLog
	docs sync.RWMutex // W protects all docs, R needs to be held when locking just one doc
	port int
	addr string

	secret []byte

	db    *mongo.Collection
	users *mongo.Collection
}

// processes a request
// called while holding s.docs lock
func (s *Server) handle(r Request) {
	doc := s.Docs[r.DocId]

	// incorrect sequence
	if r.Seq != doc.AppliedSeq[r.Uid] {
		return
	}

	ops := r.Ops

	// xform
	view := doc.Doc.View
	fmt.Println("--- HANDLE ---")
	for _, op := range r.Ops {
		fmt.Printf("%+v\n", op)
	}
	for _, op := range doc.Log[len(doc.Log)-(view-r.View):] {
		if !(r.Uid == op[0].Uid && r.Ops[0].Session == op[0].Session) {
			// xform if op is not from the same session
			ops = Xform(op, ops)
		}
	}
	doc.Doc.ApplyOps(ops)

	doc.Log = append(doc.Log, ops)
	doc.AppliedSeq[r.Uid] = r.Seq + 1

	if doc.NextSeq[r.Uid] < doc.AppliedSeq[r.Uid] {
		doc.NextSeq[r.Uid] = doc.AppliedSeq[r.Uid]
	}
}

// applies commited requests to documents and saves to db
func (s *Server) update() {
	i := 0
	for {
		time.Sleep(UpdateInterval)

		s.cl.Lock()
		tmp := s.CommitLog
		s.CommitLog = make([]Request, 0)
		s.cl.Unlock()

		if len(tmp) > 0 {
			req := make([]interface{}, len(tmp))
			for i, x := range tmp {
				req[i] = interface{}(x)
			}

			s.db.InsertMany(context.TODO(), req)

			s.docs.Lock()
			for _, r := range tmp {
				s.handle(r)
			}
			s.docs.Unlock()

			log.Println("Done")
		}

		if i%SnapMult == 0 {
			// snapshot
			s.saveToDisk()
			log.Println("Saved to disk")
		}
		i++
	}
}

func (s *Server) saveToDisk() {
	s.docs.Lock()
	defer s.docs.Unlock()

	s.cl.Lock()
	defer s.cl.Unlock()

	path := fmt.Sprintf("snapshot-%s-%d", s.addr, s.port)

	f, err := os.Create(path + ".tmp")
	if err != nil {
		log.Println("Could not create file for snapshot")
		return
	}

	res, err := json.MarshalIndent(s, "", "	")
	if err != nil {
		log.Println("Could not encode for snapshot")
		return
	}

	if _, err = f.Write(res); err != nil {
		log.Println("Could not write for snapshot")
		return
	}

	if os.Rename(path+".tmp", path+".json") != nil {
		log.Println("Could not rename file for snapshot")
		return
	}
}

func recoverFromDisk(addr string, port int) *Server {
	path := fmt.Sprintf("snapshot-%s-%d.json", addr, port)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Server{
			Docs:      make(map[int64]*DocMeta, 0),
			CommitLog: make([]Request, 0),
			addr:      addr,
			port:      port,
		}
	}

	s := new(Server)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Could not recover file")
	}

	dec := json.NewDecoder(f)
	err = dec.Decode(s)
	if err != nil {
		log.Fatal("Could not decode file")
	}

	s.addr = addr
	s.port = port

	return s
}

func (s *Server) NewClient(docId int64, uid string, conn *websocket.Conn) Client {
	return Client{
		s:     s,
		doc:   s.Docs[docId],
		conn:  conn,
		uid:   uid,
		alive: true,
	}
}

func (s *Server) recoverFromMongo() {
	MONGODB_URI := os.Getenv("MONGODB_URI")

	opt := options.Client().ApplyURI(MONGODB_URI)
	client, _ := mongo.Connect(context.TODO(), opt)

	s.db = client.Database("gopad").Collection("log")
	s.users = client.Database("gopad").Collection("users")

	// check json not ahead of db somehow
	if s.LastCommit != 0 {
		cur, err := s.db.Find(context.TODO(), bson.D{{"num", s.LastCommit - 1}})
		if err != nil {
			log.Fatal(err)
		}

		if !cur.Next(context.TODO()) {
			log.Fatal("Missing entries")
		}
	}

	cur, err := s.db.Find(context.TODO(), bson.D{{"num", bson.D{{"$gte", s.LastCommit}}}})
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var elem Request
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Mongo recovered request %d\n", elem.Num)
		s.CommitLog = append(s.CommitLog, elem)
	}

	// populate docs from log
	for _, req := range s.CommitLog {
		if _, ok := s.Docs[req.DocId]; !ok {
			s.Docs[req.DocId] = &DocMeta{
				Doc: Doc{
					Body:  []byte{},
					View:  0,
					DocId: req.DocId,
				},
				Log:        [][]Op{},
				NextSeq:    make(map[string]int, 0),
				AppliedSeq: make(map[string]int, 0),
				DocId:      req.DocId,
			}
		}

		s.handle(req)
		s.LastCommit++
	}

	s.CommitLog = make([]Request, 0)
}

func NewServer(addr string, port int) *Server {
	s := recoverFromDisk(addr, port)
	if godotenv.Load() != nil {
		log.Fatal("Error loading .env file")
	}

	s.secret = []byte(os.Getenv("JWT_SECRET"))
	log.Printf("Recovered %d log\n", s.LastCommit)

	s.recoverFromMongo()

	log.Println("Started")

	return s
}

// set up websocket
func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	docId, err := strconv.ParseInt(mux.Vars(r)["docid"], 10, 64)
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	s.docs.Lock()
	if _, ok := s.Docs[docId]; !ok {
		s.Docs[docId] = &DocMeta{
			Doc: Doc{
				Body:  []byte{},
				View:  0,
				DocId: docId,
			},

			Log:        [][]Op{},
			NextSeq:    make(map[string]int, 0),
			AppliedSeq: make(map[string]int, 0),
			DocId:      docId,
		}
	}
	s.docs.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// wait for uid message
	_, res, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	uid, ok := s.parseJWT(string(res))
	if !ok {
		conn.Close()
		return
	}

	c := s.NewClient(docId, uid, conn)
	c.interact()
}

func (s *Server) edit(w http.ResponseWriter, r *http.Request) {
	_, err := strconv.ParseInt(mux.Vars(r)["docid"], 10, 64)
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, "dist/index.html")
}
