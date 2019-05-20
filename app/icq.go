package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	mongodb "github.com/MindHunter86/icqdumper/system/mongodb"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	ICQApi struct {
		aimsid string
		client *http.Client
	}

	icqApiResponse struct {
		Timestamp int64           `json:"ts,omitempty"`
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
		LastRead        int `json:"lastRead,omitempty"`
		LastDelivered   int `json:"lastDelivered,omitempty"`
		LastReadMention int `json:"lastReadMention,omitempty"`
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
		Time       int          `json:"time,omitempty"`
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

	// POST /rapi (getHistory)
	getHistoryReq struct {
		Method string               `json:"method,omitempty"`
		ReqId  string               `json:"reqId,omitempty"`
		Aimsid string               `json:"aimsid,omitempty"`
		Params *getHistoryReqParams `json:"params,omitempty"`
	}
	getHistoryReqParams struct {
		Sn           string `json:"sn,omitempty"`
		FromMsgId    uint64 `json:"fromMsgId,omitempty"`
		Count        int    `json:"count,omitempty"`
		PatchVersion string `json:"patchVersion,omitempty"`
	}

	getHistoryRsp struct {
		Timestamp uint64 `json:"ts,omitempty"`
		//		Status    interface{}          `json:"-"`
		Method  string               `json:"method,omitempty"`
		ReqId   string               `json:"reqId,omitempty"`
		Results *getHistoryRspResult `json:"results,omitempty"`
	}

	getHistoryRspStatus struct {
		Code int `json:"code,omitempty"`
	}

	getHistoryRspResult struct {
		Messages     []*getHistoryRspResultMessage `json:"messages,omitempty"`
		LastMsgId    int                           `json:"lastMsgId,omitempty"`
		PatchVersion string                        `json:"patchVersion,omitempty"`
		Yours        *getHistoryRspResultYours     `json:"yours,omitempty"`
		Unreads      int                           `json:"ureads,omitempty"`
		UnreadCnt    int                           `json:"unreadCnt,omitempty"`
		Patch        []interface{}                 `json:"-"`
		Persons      []interface{}                 `json:"-"`
	}

	getHistoryRspResultYours struct {
		LastRead        int `json:"lastRead,omitempty"`
		LastDelivered   int `json:"lastDelivered,omitempty"`
		LastReadMention int `json:"lastReadMention,omitempty"`
	}

	getHistoryRspResultMessage struct {
		ReadsCount int                             `json:"-"`
		MsgId      uint64                          `json:"msgId,omitempty"`
		Time       int64                           `json:"time,omitempty"`
		Wid        string                          `json:"wid,omitempty"`
		Chat       *getHistoryRspResultMessageChat `json:"chat,omitempty"`
		Text       string                          `json:"text,omitempty"`
		Outgoing   bool                            `json:"-"`
		Snippets   interface{}                     `json:"-"`
	}

	getHistoryRspResultMessageChat struct {
		Sender      string      `json:"sender,omitempty"`
		Name        string      `json:"name,omitempty"`
		MemberEvent interface{} `json:"-"`
	}

	// GET /getBuddyList
	getBuddyListRsp struct {
		Response *getBuddyListRspResponse `json:response`
	}
	getBuddyListRspResponse struct {
		StatusCode int                  `json:"statusCode,omitempty"`
		StatusText string               `json:"statusText,omitempty"`
		Data       *getBuddyListRspData `json:"data,omitempty"`
	}
	getBuddyListRspData struct {
		Groups []*getBuddyListRspDataGroup `json:"gropus,omitempty"`
	}

	getBuddyListRspDataGroup struct {
		Name    string                           `json:"name,omitempty"`
		Id      int                              `json:"id,omitempty"`
		Buddies []*getBuddyListRspDataGroupBuddy `json:"buddies,omitempty"`
	}
	getBuddyListRspDataGroupBuddy struct {
		AimId     string `json:"aimId,omitempty"`
		DisplayId string `json:"displayId,omitempty"`
		Friendly  string `json:"friendly,omitempty"`
		UserType  string `json:"userType,omitempty"`
	}
)

func NewICQApi(aimsid string) (icqApi *ICQApi) {
	return &ICQApi{
		aimsid: aimsid,
		client: &http.Client{
			Timeout: 3 * time.Second,
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

func (m *ICQApi) getChats() (chats []string, e error) {

	gLogger.Debug().Msg("Trying to fetch chats...")

	var reqId uuid.UUID
	if reqId, e = uuid.NewV4(); e != nil {
		return nil, e
	}

	var reqUrl *url.URL
	if reqUrl, e = url.Parse("https://botapi.icq.net/getBuddyList?aimsid=" + m.aimsid + "&r=" + reqId.String()); e != nil {
		return nil, e
	}

	var req *http.Request
	if req, e = http.NewRequest("GET", reqUrl.String(), nil); e != nil {
		return nil, e
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/62.0.3202.89 Chrome/62.0.3202.89 Safari/537.36")
	req.Header.Set("Origin", "https://botapi.icq.net/getBuddyList")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	var rsp *http.Response
	if rsp, e = m.client.Do(req); e != nil {
		return nil, e
	}
	defer rsp.Body.Close()

	gLogger.Info().Str("response code", rsp.Status).Msg("ICQ api request has been successful")
	return m.getChatsResponse(&rsp.Body)
}

func (m *ICQApi) getChatsResponse(r *io.ReadCloser) (chats []string, e error) {
	var data []byte
	if data, e = ioutil.ReadAll(*r); e != nil {
		return nil, e
	}

	fmt.Println(string(data))

	var chatsResponse = new(getBuddyListRsp)
	if e = json.Unmarshal(data, &chatsResponse); e != nil {
		return nil, e
	}

	fmt.Println(chatsResponse.Response.StatusText)
	gLogger.Info().Msg("ICQ api request has been successfully parsed")
	return m.parseChatResponse(chatsResponse)
}

func (m *ICQApi) parseChatResponse(chatResponse *getBuddyListRsp) (chats []string, e error) {

	var chatsCollections []interface{}
	fmt.Println(chatResponse.Response.StatusCode)
	for _, v := range chatResponse.Response.Data.Groups {
		for _, v2 := range v.Buddies {
			chats = append(chats, v2.AimId)

			chatsCollections = append(chatsCollections, mongodb.CollectionChats{
				ID:    primitive.NewObjectID(),
				Name:  v2.Friendly,
				AimId: v2.AimId,
			})
		}
	}

	if e = gMongoDB.InsertMany("chats", &chatsCollections); e != nil {
		return nil, e
	}

	return nil, e
}

func (m *ICQApi) getChatsMessages(chatIds []string) (e error) {
	for _, v := range chatIds {
		if e = m.getChatMessages(v, 1); e != nil {
			break
		}
	}

	return e
}

func (m *ICQApi) getChatMessages(chatId string, fromMsgId uint64) (e error) {

	gLogger.Debug().Str("chatId", chatId).Uint64("lastMsgId", fromMsgId).Msg("Trying to fetch messages for chat")

	var reqId uuid.UUID
	if reqId, e = uuid.NewV4(); e != nil {
		return e
	}

	var reqUrl *url.URL
	if reqUrl, e = url.Parse("https://botapi.icq.net/rapi"); e != nil {
		return e
	}

	var buf = new(bytes.Buffer)
	if e = json.NewEncoder(buf).Encode(&getHistoryReq{
		"getHistory", reqId.String(), m.aimsid, &getHistoryReqParams{
			chatId, fromMsgId, 100, "init",
		},
	}); e != nil {
		return e
	}

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

	if rsp.StatusCode != 200 {
		return errors.New("ICQ api send non 200 OK")
	}

	var messagesResponse *getHistoryRsp
	if messagesResponse, e = m.getChatMessagesResponse(&rsp.Body); e != nil {
		return e
	}

	var lastMsgId uint64
	if lastMsgId, e = m.parseChatMessagesResponse(chatId, messagesResponse.Results.Messages); e != nil || lastMsgId == 0 {
		return e
	}

	// recursive calls
	if e = m.getChatMessages(chatId, lastMsgId); e != nil {
		return e
	}

	return e
}

func (m *ICQApi) getChatMessagesResponse(r *io.ReadCloser) (messagesResponse *getHistoryRsp, e error) {

	var data []byte
	if data, e = ioutil.ReadAll(*r); e != nil {
		return nil, e
	}

	if e = json.Unmarshal(data, &messagesResponse); e != nil {
		return nil, e
	}

	return messagesResponse, e
}

func (m *ICQApi) parseChatMessagesResponse(chatId string, messages []*getHistoryRspResultMessage) (lastMsgId uint64, e error) {

	// if no messages - exit
	if len(messages) == 0 {
		return 0, e
	}

	lastMsgId = messages[len(messages)-1].MsgId
	var chatMessages []*mongodb.CollectionChatsMessage
	for _, v := range messages {
		chatMessages = append(chatMessages, &mongodb.CollectionChatsMessage{
			MsgId:  v.MsgId,
			Time:   time.Unix(v.Time, 0),
			Wid:    v.Wid,
			Sender: v.Chat.Sender,
			Text:   v.Text,
		})
	}

	if e = gMongoDB.UpdateOne("chats", bson.M{
		"aimId": chatId,
	}, bson.M{
		"$push": bson.M{
			"aimId.$.messages": chatMessages,
		},
	}); e != nil {
		return 0, e
	}

	return lastMsgId, e
}
