package open_wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/owen-gxz/open-wechat/core"
	"net/http"
	"time"
)

const (
	componentAccessTokenUrl = wechatApiUrl + "/cgi-bin/component/api_component_token"
)

type AccessTokenServer interface {
	Token() (token string, err error)
}

type DefaultAccessTokenServer struct {
	AppID     string
	AppSecret string
	ticket    TicketServer

	ComponentAccessToken string `json:"component_access_token"`
	ExpiresIn            int64  `json:"expires_in"` // 当前时间 + 过期时间
}

// token不使用不获取
func (d *DefaultAccessTokenServer) Token() (token string, err error) {
	timeUnix := time.Now().Unix()
	if d.ExpiresIn <= time.Now().Unix() {
		ticket, err := d.ticket.GetTicket()
		if err != nil {
			return "", nil
		}
		aresp, err := GetAccessToken(d.AppID, d.AppSecret, ticket)
		if err != nil {
			return "", nil
		}
		d.ExpiresIn = timeUnix + aresp.ExpiresIn
		d.ComponentAccessToken = aresp.ComponentAccessToken
	}
	return d.ComponentAccessToken, nil
}

type AccessTokenResponse struct {
	core.Error
	ComponentAccessToken string `json:"component_access_token"`
	ExpiresIn            int64  `json:"expires_in"`
}

type AccessTokenRequest struct {
	ComponentAppid        string `json:"component_appid"`
	ComponentAppsecret    string `json:"component_appsecret"`
	ComponentVerifyTicket string `json:"component_verify_ticket"`
}

//// todo 获取第三方应用token, 该方法如果调用可能会将之前的token失效，所以取消使用
//func (srv *Server) GetAccessToken() (*AccessTokenResponse, error) {
//	ticket, err := srv.ticketServer.GetTicket()
//	if err != nil {
//		return nil, nil
//	}
//	return GetAccessToken(srv.cfg.AppID, srv.cfg.AppSecret, ticket, srv.Client)
//}

// 获取第三方应用token
func GetAccessToken(appid, AppSecret, ticket string) (*AccessTokenResponse, error) {
	req := AccessTokenRequest{
		ComponentAppid:        appid,
		ComponentAppsecret:    AppSecret,
		ComponentVerifyTicket: ticket,
	}
	resp := &AccessTokenResponse{}
	// todo
	err := postJson(componentAccessTokenUrl, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// todo
func postJson(incompleteURL string, request interface{}, response interface{}) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(&request); err != nil {
		return err
	}
	httpResp, err := http.DefaultClient.Post(incompleteURL, "application/json; charset=utf-8", &buf)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("http.Status: %s", httpResp.Status)
	}
	return json.NewDecoder(httpResp.Body).Decode(&response)
}
