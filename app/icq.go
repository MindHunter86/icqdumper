package app

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"
)

type (
	ICQApi struct {
		aimsid string
		client *http.Client
	}

	icqApiResponse struct {
		Timestamp uint64          `json:"ts,omitempty"`
		Status    *responseStatus `json:"status,omitempty"`
		Method    string          `json:"method,omitempty"`
		ReqId     string          `json:"reqId,omitempty"`
		Results   *responseResult `json:"results,omitempty"`
	}

	responseStatus struct {
		Code int `json:"code,omitempty"`
	}

	responseResult struct {
		Messages     []*resultMessage `json:"messages,omitempty"`
		LastMsgId    int              `json:"lastMsgId,omitempty"`
		PatchVersion string           `json:"patchVersion,omitempty"`
		Yours        *resultYours     `json:"yours,omitempty"`
		Unreads      int              `json:"ureads,omitempty"`
		UnreadCnt    int              `json:"unreadCnt,omitempty"`
		Patch        []*resultPatch   `json:"patch,omitempty"`
		Persons      []*resultPerson  `json:"persons,omitempty"`
	}

	resultYours struct {
		LastRead        uint64 `json:"lastRead,omitempty"`
		LastDelivered   uint64 `json:"lastDelivered,omitempty"`
		LastReadMention uint64 `json:"lastReadMention,omitempty"`
	}

	resultPatch struct {
		MsgId   uint64 `json:"msgId,omitempty"`
		Pa_type string `json:"type,omitempty"`
	}

	resultPerson struct {
		Sn        string `json:"sn,omitempty"`
		Role      string `json:"role,omitempty"`
		Friendly  string `json:"friendly,omitempty"`
		FirstName string `json:"firstName,omitempty"`
		LastName  string `json:"lastName,omitempty"`
		Nick      string `json:"nick,omitempty"`
		NickName  string `json:"nickname,omitempty"`
	}

	resultMessage struct {
		ReadsCount int          `json:"-"`
		MsgId      uint64       `json:"msgId,omitempty"`
		Time       uint64       `json:"time,omitempty"`
		Wid        string       `json:"wid,omitempty"`
		Chat       *messageChat `json:"chat,omitempty"`
		Text       string       `json:"text,omitempty"`
		Outgoing   bool         `json:"-"`
		Snippets   interface{}  `json:"-"`
	}

	messageChat struct {
		Sender      string      `json:"sender,omitempty"`
		Name        string      `json:"name,omitempty"`
		MemberEvent interface{} `json:"-"`
	}

	chatMemberEvent struct {
		Ev_type string   `json:"type,omitempty"`
		Role    string   `json:"role,omitempty"`
		Members []string `json:"members,omitempty"`
	}

	icqRequest struct {
		Method string         `json:"method,omitempty"`
		ReqId  string         `json:"reqId,omitempty"`
		Aimsid string         `json:"aimsid,omitempty"`
		Params *requestParams `json:"params,omitempty"`
	}
	requestParams struct {
		Sn           string `json:"sn,omitempty"`
		FromMsgId    uint64 `json:"fromMsgId,omitempty"`
		Count        int    `json:"count,omitempty"`
		PatchVersion string `json:"patchVersion,omitempty"`
	}
)

func NewICQApi(aimsid string) *ICQApi {
	return &ICQApi{
		aimsid: aimsid,
		client: &http.Client{
			Timeout: time.Second * 3, // todo - add this to cli.Flags
		},
	}
}

func (m *ICQApi) dumpHistroyFromChat(chatId string) (e error) {
	var lastMsgId uint64 = 1

	gLogger.Debug().Msg("Start dumpHistroy loop")
	for e == nil {
		lastMsgId, e = m.dumpPartialHistroy(chatId, lastMsgId)
	}

	return e
}

func (m *ICQApi) dumpPartialHistroy(chatId string, fromMsgId uint64) (lastMsgId uint64, e error) {

	gLogger.Debug().Str("chatId", chatId).Uint64("fromMsgId", fromMsgId).Msg("Start ICQ API reqeust builder")

	var reqUrl *url.URL
	if reqUrl, e = url.Parse("https://botapi.icq.net/rapi"); e != nil {
		return 0, e
	}

	var buf = new(bytes.Buffer)

	var reqId uuid.UUID
	if reqId, e = uuid.NewV4(); e != nil {
		return 0, e
	}

	var reqBodyParams = &requestParams{
		chatId, fromMsgId, 10, "init",
	}
	var reqBody = &icqRequest{
		"getHistory", reqId.String(), m.aimsid,
		reqBodyParams,
	}

	if e = json.NewEncoder(buf).Encode(reqBody); e != nil {
		return 0, e
	}

	var req *http.Request
	if req, e = http.NewRequest("POST", reqUrl.String(), buf); e != nil {
		return 0, e
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/62.0.3202.89 Chrome/62.0.3202.89 Safari/537.36")
	req.Header.Set("Origin", "https://botapi.icq.net/rapi")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	var rsp *http.Response
	if rsp, e = m.client.Do(req); e != nil {
		return 0, e
	}
	defer rsp.Body.Close()

	return m.parseApiResponse(&rsp.Body)
}

func (m *ICQApi) parseApiResponse(r *io.ReadCloser) (lastMsgId uint64, e error) {

	var data []byte
	if data, e = ioutil.ReadAll(*r); e != nil {
		return 0, e
	}

	var apiResponse = new(icqApiResponse)
	if e = json.Unmarshal(data, apiResponse); e != nil {
		return 0, e
	}

	m.apiResultDebug(apiResponse)
	return m.getLastMessageFromResponse(apiResponse.Results.Messages), e
}

func (m *ICQApi) getLastMessageFromResponse(messages []*resultMessage) (lastMsgId uint64) {
	lastMsgId = messages[len(messages)-1].MsgId
	return lastMsgId
}

func (m *ICQApi) apiResultDebug(res *icqApiResponse) {
	for i := 0; i < len(res.Results.Messages); i++ {
		gLogger.Info().Str("author", res.Results.Messages[i].Chat.Sender).Str("msg", res.Results.Messages[i].Text).Uint64("mid", res.Results.Messages[i].MsgId).Msg("message parsed")
	}
}
