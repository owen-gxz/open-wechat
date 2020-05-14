package open_wechat

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/owen-gxz/open-wechat/util"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

// open api 配置
type Config struct {
	AppID     string
	AppSecret string
	AESKey    string
	Token     string
	//RedirectUrl    string
}
type HandlerChain func(c Context)

type Server struct {
	sync.Mutex
	cfg          Config
	handlerMap   map[string]HandlerChain //方法处理
	DecodeAesKey []byte
	*Client
	errorHandler WechatErrorer           // 错误处理
	ticketServer TicketServer // ticket存储
	// 获取token
	AccessTokenServer

}

const (
	// InfoTypeVerifyTicket 返回ticket
	InfoTypeVerifyTicket = "component_verify_ticket"
	// InfoTypeAuthorized 授权
	InfoTypeAuthorized = "authorized"
	// InfoTypeUnauthorized 取消授权
	InfoTypeUnauthorized = "unauthorized"
	// InfoTypeUpdateAuthorized 更新授权
	InfoTypeUpdateAuthorized = "updateauthorized"

	wechatApiUrl = "https://api.weixin.qq.com"
)

func (srv *Server) getAESKey() []byte {
	return srv.DecodeAesKey
}
func (srv *Server) getToken() string {
	return srv.cfg.Token
}

type cipherRequestHttpBody struct {
	XMLName            struct{} `xml:"xml"`
	ToUserName         string   `xml:"ToUserName"`
	AppId              string   `xml:"AppId"` // openapi use
	Base64EncryptedMsg []byte   `xml:"Encrypt"`
}

func NewService(cfg Config, ticket TicketServer, cli *http.Client, tokenService AccessTokenServer, errHandler WechatErrorer) *Server {
	if errHandler == nil {
		errHandler = DefaultErrorHandler
	}
	if ticket == nil {
		ticket = defaultTicketServerHander
	}
	client := NewClient(cli)
	if tokenService == nil {
		tokenService = &DefaultAccessTokenServer{ticket: ticket, AppID: cfg.AppID, AppSecret: cfg.AppSecret}
	}
	srv := Server{
		cfg:               cfg,
		errorHandler:      errHandler,
		handlerMap:        make(map[string]HandlerChain),
		ticketServer:      ticket,
		Client:            client,
		AccessTokenServer: tokenService,
	}
	srv.Lock()
	//todo  用户是可以覆盖的
	srv.AddHander(InfoTypeVerifyTicket, func(c Context) {
		err := srv.ticketServer.SetTicket(c.MixedMsg.ComponentVerifyTicket)
		if err != nil {
			srv.errorHandler.ServeError(c.w, c.r, err)
		}
		c.w.Write(Success)
	})
	defer srv.Unlock()
	if cfg.AESKey != "" {
		if len(cfg.AESKey) != 43 {
			panic("the length of base64AESKey must equal to 43")
		}
		var err error
		srv.DecodeAesKey, err = base64.StdEncoding.DecodeString(cfg.AESKey + "=")
		if err != nil {
			panic(fmt.Sprintf("Decode base64AESKey:%s failed", cfg.AESKey))
		}
	}
	return &srv
}

func (srv *Server) AddHander(t string, hander HandlerChain) {
	srv.handlerMap[t] = hander
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query:=r.URL.Query()

	switch r.Method {
	case "POST": // 推送消息(事件)
		switch encryptType := query.Get("encrypt_type"); encryptType {
		case "aes":
			haveSignature := query.Get("signature")
			if haveSignature == "" {
				srv.errorHandler.ServeError(w, r, errors.New("not found signature query parameter"))
				return
			}
			haveMsgSignature := query.Get("msg_signature")
			if haveMsgSignature == "" {
				srv.errorHandler.ServeError(w, r, errors.New("not found msg_signature query parameter"))
				return
			}
			timestampString := query.Get("timestamp")
			if timestampString == "" {
				srv.errorHandler.ServeError(w, r, errors.New("not found timestamp query parameter"))
				return
			}
			_, err := strconv.ParseInt(timestampString, 10, 64)
			if err != nil {
				err = fmt.Errorf("can not parse timestamp query parameter %q to int64", timestampString)
				srv.errorHandler.ServeError(w, r, err)
				return
			}
			nonce := query.Get("nonce")
			if nonce == "" {
				srv.errorHandler.ServeError(w, r, errors.New("not found nonce query parameter"))
				return
			}

			var token string
			currentToken := srv.getToken()
			if currentToken == "" {
				err = errors.New("token was not set for Server, see NewServer function or Server.SetToken method")
				srv.errorHandler.ServeError(w, r, err)
				return
			}
			token = currentToken
			wantSignature := util.Sign(token, timestampString, nonce)
			if haveSignature != wantSignature {
				return
			}
			requestHttpBody := cipherRequestHttpBody{}
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				srv.errorHandler.ServeError(w, r, err)
				return
			}
			err = xml.Unmarshal(data, &requestHttpBody)
			if err != nil {
				srv.errorHandler.ServeError(w, r, err)
				return
			}

			wantMsgSignature := util.MsgSign(token, timestampString, nonce, string(requestHttpBody.Base64EncryptedMsg))
			if haveMsgSignature!=wantMsgSignature {
				err = fmt.Errorf("check msg_signature failed, have: %s, want: %s", haveMsgSignature, wantMsgSignature)
				srv.errorHandler.ServeError(w, r, err)
				return
			}

			encryptedMsg := make([]byte, base64.StdEncoding.DecodedLen(len(requestHttpBody.Base64EncryptedMsg)))
			encryptedMsgLen, err := base64.StdEncoding.Decode(encryptedMsg, requestHttpBody.Base64EncryptedMsg)
			if err != nil {
				srv.errorHandler.ServeError(w, r, err)
				return
			}
			encryptedMsg = encryptedMsg[:encryptedMsgLen]

			var aesKey []byte
			aesKey = srv.getAESKey()
			if aesKey == nil {
				err = errors.New("aes key was not set for Server, see NewServer function or Server.SetAESKey method")
				srv.errorHandler.ServeError(w, r, err)
				return
			}
			_, msgPlaintext, haveAppIdBytes, err := util.AESDecryptMsg(encryptedMsg, aesKey)
			if err != nil {
				return
			}

			haveAppId := string(haveAppIdBytes)
			wantAppId := srv.cfg.AppID
			if wantAppId != "" && haveAppId!=wantAppId {
				err = fmt.Errorf("the message AppId mismatch, have: %s, want: %s", haveAppId, wantAppId)
				srv.errorHandler.ServeError(w, r, err)
				return
			}
			var mixedMsg MixedMsg
			if err = xml.Unmarshal(msgPlaintext, &mixedMsg); err != nil {
				srv.errorHandler.ServeError(w, r, err)
				return
			}
			ctx := Context{
				w:             w,
				r:             r,
				MsgCiphertext: encryptedMsg,
				MsgPlaintext:  msgPlaintext,
				MixedMsg:      &mixedMsg,
			}
			hand, exit := srv.handlerMap[mixedMsg.InfoType]
			if !exit {
				srv.errorHandler.ServeError(w, r, errors.New("no hander"))
				return
			}
			hand(ctx)
		default:
			srv.errorHandler.ServeError(w, r, errors.New("unknown encrypt_type: "+encryptType))
		}
	case "GET": // 验证回调URL是否有效
		haveSignature := query.Get("signature")
		if haveSignature == "" {
			srv.errorHandler.ServeError(w, r, errors.New("not found signature query parameter"))
			return
		}
		timestamp := query.Get("timestamp")
		if timestamp == "" {
			srv.errorHandler.ServeError(w, r, errors.New("not found timestamp query parameter"))
			return
		}
		nonce := query.Get("nonce")
		if nonce == "" {
			srv.errorHandler.ServeError(w, r, errors.New("not found nonce query parameter"))
			return
		}
		echostr := query.Get("echostr")
		if echostr == "" {
			srv.errorHandler.ServeError(w, r, errors.New("not found echostr query parameter"))
			return
		}

		var token string
		token = srv.getToken()
		wantSignature := util.Sign(token, timestamp, nonce)
		if haveSignature != wantSignature {
			srv.errorHandler.ServeError(w, r, errors.New("sign error"))
			return
		}
		io.WriteString(w, echostr)
	}
}

//
