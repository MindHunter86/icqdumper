package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"
)

type (
	ICQApi struct {
		Aimsid string
		Client *http.Client
	}

	icqApiResponse struct {
		Timestamp uint64            `json:"ts"`
		Status    *responseStatus   `json:"status"`
		Method    string            `json:"method"`
		ReqId     string            `json:"reqid"`
		Results   []*responseResult `json:"results"`
	}

	responseStatus struct {
		Code int `json:"code"`
	}

	responseResult struct {
		Messages     []*resultMessage `json:"messages"`
		LastMsgId    uint64           `json:"lastmsgid"`
		PatchVersion string           `json:"patchversion"`
		Yours        *resultYours     `json:"yours"`
		Unreads      uint             `json:"unreads"`
		UnreadCnt    uint             `json:"unreadcnt"`
		Path         []*resultPatch   `json:"path"`
		Persons      []*resultPerson  `json:"persons"`
	}

	resultYours struct {
		LastRead        uint64 `json:"lastread"`
		LastDelivered   uint64 `json:"lastdelivered"`
		LastReadMention uint64 `json:"lastreadmention"`
	}

	resultPatch struct {
		MsgId   uint64 `json:"msgid"`
		Pa_type string `json:"type"`
	}

	resultPerson struct {
		Sn        string `json:"sn"`
		Role      string `json:"role"`
		Friendly  string `json:"friendly"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	}

	resultMessage struct {
		MsgId uint64       `json:"msgid"`
		Time  uint64       `json:"time"`
		Wid   string       `json:"wid"`
		Chat  *messageChat `json:"chat"`
		Text  string       `json:"text"`
	}

	messageChat struct {
		Senders     string           `json:"senders"`
		Name        string           `json:"name"`
		MemberEvent *chatMemberEvent `json:"memberevent"`
	}

	chatMemberEvent struct {
		Ev_type string   `json:"type"`
		Role    string   `json:"role"`
		Members []string `json:"members"`
	}

	icqRequest struct {
		Method string         `json:"method"`
		ReqId  string         `json:"reqId"`
		Aimsid string         `json:"aimsid"`
		Params *requestParams `json:"params"`
	}
	requestParams struct {
		Sn           string `json:"sn"`
		FromMsgId    int    `json:"fromMsgId"`
		Count        int    `json:"count"`
		PatchVersion string `json:"patchVersion"`
	}
)

func NewICQApi(aimsid string) *ICQApi {
	return &ICQApi{
		aimsid: aimsid,
		client: &http.Client{
			Timeout: time.Second * 1, // todo - add this to cli.Flags
		},
	}
}

func (m *ICQApi) dumpHistroyFromChat(chatId string) (e error) {

	var reqUrl *url.URL
	if reqUrl, e = url.Parse("https://botapi.icq.net/rapi"); e != nil {
		return e
	}

	var buf = new(bytes.Buffer)

	var reqId uuid.UUID
	if reqId, e = uuid.NewV4(); e != nil {
		return e
	}

	var reqBodyParams = &requestParams{
		chatId, 6369479187785847000 - 1, 100, "init",
	}
	var reqBody = &icqRequest{
		"getHistory", reqId.String(), m.aimsid,
		reqBodyParams,
	}

	if e = json.NewEncoder(buf).Encode(reqBody); e != nil {
		return e
	}

	fmt.Println(buf.String())

	var req *http.Request
	if req, e = http.NewRequest("POST", reqUrl.String(), buf); e != nil {
		return e
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/62.0.3202.89 Chrome/62.0.3202.89 Safari/537.36")
	req.Header.Set("Origin", "https://botapi.icq.net/rapi")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	var rsp *http.Response
	if rsp, e = m.client.Do(req); e != nil {
		return e
	}
	defer rsp.Body.Close()

	if tmp, e := ioutil.ReadAll(rsp.Body); e != nil {
		return e
	} else {
		gLogger.Info().Str("body", string(tmp)).Msg("New response from ICQ API!")
	}

	return e
}
