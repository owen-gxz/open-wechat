package open_wechat

import "net/http"

type Context struct {
	w http.ResponseWriter
	r *http.Request

	MsgCiphertext []byte    // 消息的密文文本
	MsgPlaintext  []byte    // 消息的明文文本, xml格式
	MixedMsg      *MixedMsg // 消息
}
