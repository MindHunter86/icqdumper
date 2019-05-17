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
		Timestamp uint64 `json:"ts"`
		Status    struct {
			Code int
		}
		Method  string
		ReqId   string
		Results []*responseResult
	}

	responseResult struct {
		Messages     []*resultMessage
		LastMsgId    uint64
		PatchVersion string
		Yours        *resultYours
		Unreads      uint
		UnreadCnt    uint
		Path         []*resultPatch
		Persons      []*resultPerson
	}

	resultYours struct {
		LastRead        uint64
		LastDelivered   uint64
		LastReadMention uint64
	}

	resultPatch struct {
		MsgId   uint64
		Pa_type string `json:"type"`
	}

	resultPerson struct {
		Sn        string
		Role      string
		Friendly  string
		FirstName string
		LastName  string
	}

	resultMessage struct {
		MsgId uint64
		Time  uint64
		Wid   string
		Chat  *messageChat
		Text  string
	}

	messageChat struct {
		Senders     string
		Name        string
		MemberEvent *chatMemberEvent
	}

	chatMemberEvent struct {
		Ev_type string `json:"type"`
		Role    string
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
