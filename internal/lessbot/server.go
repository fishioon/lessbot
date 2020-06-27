package lessbot

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/fishioon/lessbot/internal/wxbiz"
)

var (
	ErrSystemError         = NewError(1, "system error")
	ErrBadCommand          = NewError(2, "bad command")
	ErrInvalidDays         = NewError(3, "bad amount")
	ErrInsufficientBalance = NewError(4, "insufficient balance")
	ErrPermissionDenied    = NewError(5, "permission denied")
)

type YoErr struct {
	Code    int
	Message string
}

type Bot struct {
	AppID     string
	AppSecret string
	wx        *wxbiz.Wxbiz
}

func (e *YoErr) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

func NewError(code int, message string) *YoErr {
	return &YoErr{code, message}
}

type Server struct {
	mux  *http.ServeMux
	bots map[string]*Bot
}

func NewServer() (*Server, error) {
	s := &Server{
		mux: http.NewServeMux(),
	}

	s.init()
	return s, nil
}

type response struct {
	Errors string      `json:"errors"`
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) init() {
	// init all bots

	s.mux.HandleFunc("/wxbot/listen/", s.handleListen)
}

func (s *Server) handleListen(w http.ResponseWriter, r *http.Request) {
	appid := strings.TrimPrefix(r.URL.Path, "/wxbot/listen/")
	bot, ok := s.bots[appid]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	sign := r.URL.Query().Get("msg_signature")
	ts := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	if r.Method == "GET" {
		// check url
		echostr := r.URL.Query().Get("echostr")
		msg, err := bot.wx.VerifyURL(sign, ts, nonce, echostr)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Write([]byte(msg))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	recv, err := bot.wx.Decode(sign, ts, nonce, string(body))
	if err != nil {
		log.Printf("Error parse msg: %v", err)
		http.Error(w, "invalid msg", http.StatusBadRequest)
		return
	}
	reply, err := s.botExec(bot, []byte(recv))
	if err != nil {
		log.Printf("Error parse msg: %v", err)
		http.Error(w, "invalid msg", http.StatusBadRequest)
		return
	}
	res, err := bot.wx.Encode(ts, nonce, string(reply))
	if err != nil {
		log.Printf("Error pack msg: %v", err)
		http.Error(w, "pack msg error", http.StatusBadRequest)
		return
	}
	log.Printf("reply: %s", res)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(res))
}

func (s *Server) botExec(bot *Bot, msg []byte) ([]byte, error) {
	return nil, nil
}
