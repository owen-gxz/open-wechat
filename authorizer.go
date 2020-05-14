package open_wechat

import "github.com/owen-gxz/open-wechat/core"

// 授权方信息
const (
	AuthorizerInfoUrl      = wechatApiUrl + "/cgi-bin/component/api_get_authorizer_info?component_access_token=%s"
	AuthorizerOptionUrl    = wechatApiUrl + "/cgi-bin/component/api_get_authorizer_option?component_access_token=%s"
	SetAuthorizerOptionUrl = wechatApiUrl + "/cgi-bin/component/api_set_authorizer_option?component_access_token=%s"
	AuthorizerListUrl      = wechatApiUrl + "/cgi-bin/component/api_get_authorizer_list?component_access_token=%s"
)

type AuthorizerInfo struct {
}
type AuthorizerInfoRequest struct {
	ComponentAppid  string `json:"component_appid"`
	AuthorizerAppid string `json:"authorizer_appid"`
}

type AuthorizerInfoResponse struct {
	core.Error
	ComponentAppid  string `json:"component_appid"`
	AuthorizerAppid string `json:"authorizer_appid"`
}

// 获取授权法信息
func (srv Server) AuthorizerInfo(authorizerAppid string) (*AuthorizerInfoResponse, error) {
	accessToken, err := srv.Token()
	if err != nil {
		return nil, err
	}
	req := AuthorizerInfoRequest{
		ComponentAppid:  srv.cfg.AppID,
		AuthorizerAppid: authorizerAppid,
	}
	resp := AuthorizerInfoResponse{}
	err = srv.PostJson(getCompleteUrl(AuthorizerInfoUrl, accessToken), req, resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type AuthorizeOption string

// option类型
var (
	AuthorizeOptionLocal           AuthorizeOption = "location_report"
	AuthorizeOptionVoiceRecognize  AuthorizeOption = "voice_recognize"
	AuthorizeOptionCustomerService AuthorizeOption = "customer_service"
)

type AuthorizerOptionRequest struct {
	ComponentAppid  string          `json:"component_appid"`
	AuthorizerAppid string          `json:"authorizer_appid"`
	OptionName      AuthorizeOption `json:"option_name"`
}

type AuthorizerOptionResponse struct {
	core.Error
	AuthorizerAppid string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// 获取选项信息
func (srv Server) AuthorizerOption(authorizerAppid string, optionName AuthorizeOption) (*AuthorizerOptionResponse, error) {
	accessToken, err := srv.Token()
	if err != nil {
		return nil, err
	}
	req := AuthorizerOptionRequest{
		ComponentAppid:  srv.cfg.AppID,
		AuthorizerAppid: authorizerAppid,
		OptionName:      optionName,
	}
	resp := AuthorizerOptionResponse{}
	err = srv.PostJson(getCompleteUrl(AuthorizerOptionUrl, accessToken), req, resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetAuthorizerOptionRequest struct {
	AuthorizerOptionRequest
	OptionValue string `json:"option_name"`
}

type SetAuthorizerOptionResponse struct {
	core.Error
}

// 设置选项信息
func (srv Server) SetAuthorizerOption(authorizerAppid string, optionName AuthorizeOption, optionValue string) (*SetAuthorizerOptionResponse, error) {
	accessToken, err := srv.Token()
	if err != nil {
		return nil, err
	}
	req := SetAuthorizerOptionRequest{
		AuthorizerOptionRequest: AuthorizerOptionRequest{
			ComponentAppid:  srv.cfg.AppID,
			AuthorizerAppid: authorizerAppid,
			OptionName:      optionName,
		},
		OptionValue: optionValue,
	}
	resp := SetAuthorizerOptionResponse{}
	err = srv.PostJson(getCompleteUrl(SetAuthorizerOptionUrl, accessToken), req, resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type AuthorizerListRequest struct {
	ComponentAppid string `json:"component_appid"`
	Offset         int    `json:"offset"`
	Count          int    `json:"count"`
}

type AuthorizerListResponse struct {
	core.Error
	TotalCount int `json:"total_count"`
	List       []struct {
		AuthorizerAppid string `json:"authorizer_appid"`
		RefreshToken    string `json:"refresh_token"`
		AuthTime        int    `json:"auth_time"`
	} `json:"list"`
}

// 拉取用户授权列表
func (srv Server) AuthorizerList(offset, count int) (*AuthorizerListResponse, error) {
	accessToken, err := srv.Token()
	if err != nil {
		return nil, err
	}
	req := AuthorizerListRequest{
		ComponentAppid: srv.cfg.AppID,
		Offset:         offset,
		Count:          count,
	}
	resp := AuthorizerListResponse{}
	err = srv.PostJson(getCompleteUrl(AuthorizerListUrl, accessToken), req, resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
