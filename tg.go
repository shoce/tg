/*
history:
025/0404 v1

GoFmt
GoBuildNull
GoRun
*/

package tg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	NL = "\n"

	ApiUrlBase = "https://api.telegram.org"

	// https://core.telegram.org/bots/api#formatting-options
	ParseMode = "MarkdownV2"
)

var (
	DEBUG bool

	HttpClient = &http.Client{}
)

func Esc(text string) string {
	// https://core.telegram.org/bots/api#formatting-options
	for _, c := range "\\_*[]()~`>#+-=|{}.!" {
		text = strings.ReplaceAll(text, string(c), "\\"+string(c))
	}
	return text
}

func Bold(text string) string {
	// https://core.telegram.org/bots/api#formatting-options
	return "*" + Esc(text) + "*"
}

type Document struct {
	// https://core.telegram.org/bots/api#document
}

type VideoNote struct {
	// https://core.telegram.org/bots/api#videonote
}

type Voice struct {
	// https://core.telegram.org/bots/api#voice
}

type Location struct {
	// https://core.telegram.org/bots/api#location
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	HorizontalAccuracy   float64 `json:"horizontal_accuracy"`
	LivePeriod           int64   `json:"live_period"`
	Heading              int64   `json:"heading"`
	ProximityAlertRadius int64   `json:"proximity_alert_radius"`
}

type Message struct {
	// https://core.telegram.org/bots/api#message
	Id        string
	MessageId int64  `json:"message_id"`
	From      User   `json:"from,omitempty"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`

	Audio     Audio       `json:"audio,omitempty"`
	Document  Document    `json:"document,omitempty"`
	Photo     []PhotoSize `json:"photo,omitempty"`
	Video     Video       `json:"video,omitempty"`
	VideoNote VideoNote   `json:"video_note,omitempty"`
	Voice     Voice       `json:"voice,omitempty"`

	Caption               string `json:"caption,omitempty"`
	ShowCaptionAboveMedia *bool  `json:"show_caption_above_media,omitempty"`

	Location Location `json:"location,omitempty"`
}

type User struct {
	Id        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type Chat struct {
	Id         int64  `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	InviteLink string `json:"invite_link"`
}

type PhotoSize struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
	Width        int64  `json:"width"`
	Height       int64  `json:"height"`
	FileSize     int64  `json:"file_size"`
}

type Audio struct {
	FileId       string    `json:"file_id"`
	FileUniqueId string    `json:"file_unique_id"`
	Duration     int64     `json:"duration"`
	Performer    string    `json:"performer"`
	Title        string    `json:"title"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	Thumb        PhotoSize `json:"thumb"`
}

type Video struct {
	FileId       string    `json:"file_id"`
	FileUniqueId string    `json:"file_unique_id"`
	Width        int64     `json:"width"`
	Height       int64     `json:"height"`
	Duration     int64     `json:"duration"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	Thumb        PhotoSize `json:"thumb"`
}

type Response struct {
	Ok          bool     `json:"ok"`
	Description string   `json:"description"`
	Result      *Message `json:"result"`
}

type LinkPreviewOptions struct {
	IsDisabled bool `json:"is_disabled"`
}

type SendMessageRequest struct {
	// https://core.telegram.org/bots/api#sendmessage

	ChatId           string `json:"chat_id"`
	MessageId        int64  `json:"message_id"`
	ReplyToMessageId int64  `json:"reply_to_message_id"`
	Text             string `json:"text"`
	ParseMode        string `json:"parse_mode,omitempty"`

	DisableNotification bool `json:"disable_notification,omitempty"`

	LinkPreviewOptions LinkPreviewOptions `json:"link_preview_options,omitempty"`
}

type SendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		MessageId int64 `json:"message_id"`
	} `json:"result"`
}

func SendMessage(tgtoken string, req SendMessageRequest) (msg *Message, err error) {
	// https://core.telegram.org/bots/api#sendmessage

	if DEBUG {
		log("DEBUG req==%#v", req)
	}
	if req.ParseMode == "" {
		req.ParseMode = ParseMode
	}
	reqjson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var resp Response
	err = postJson(
		fmt.Sprintf("%s/bot%s/sendMessage", ApiUrlBase, tgtoken),
		bytes.NewBuffer(reqjson),
		&resp,
	)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("sendMessage: %s", resp.Description)
	}

	msg = resp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	return msg, nil
}

type SendPhotoRequest struct {
	ChatId    string `json:"chat_id"`
	Photo     string `json:"photo"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

func SendPhoto(tgtoken string, req SendPhotoRequest) (msg *Message, err error) {
	// https://core.telegram.org/bots/api#sendphoto

	if DEBUG {
		log("DEBUG req==%#v", req)
	}
	if req.ParseMode == "" {
		req.ParseMode = ParseMode
	}
	reqjson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var resp Response
	err = postJson(
		fmt.Sprintf("%s/bot%s/sendPhoto", ApiUrlBase, tgtoken),
		bytes.NewBuffer(reqjson),
		&resp,
	)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("sendPhoto: %s", resp.Description)
	}

	msg = resp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	return msg, nil
}

type DeleteMessageRequest struct {
	ChatId    int64 `json:"chat_id"`
	MessageId int64 `json:"message_id"`
}

func DeleteMessage(tgtoken string, req DeleteMessageRequest) error {
	// https://core.telegram.org/bots/api#deletemessage

	reqjson, err := json.Marshal(req)
	if err != nil {
		return err
	}

	var resp Response
	err = postJson(
		fmt.Sprintf("%s/bot%s/deleteMessage", ApiUrlBase, tgtoken),
		bytes.NewBuffer(reqjson),
		&resp,
	)
	if err != nil {
		return fmt.Errorf("postJson: %w", err)
	}

	if !resp.Ok {
		return fmt.Errorf("deleteMessage: %s", resp.Description)
	}

	return nil
}

type PromoteChatMemberRequest struct {
	ChatId int64 `json:"chat_id"`
	UserId int64 `json:"user_id"`

	IsAnonymous         bool `json:"is_anonymous"`
	CanManageChat       bool `json:"can_manage_chat"`
	CanPostMessages     bool `json:"can_post_messages"`
	CanEditMessages     bool `json:"can_edit_messages"`
	CanDeleteMessages   bool `json:"can_delete_messages"`
	CanChangeInfo       bool `json:"can_change_info"`
	CanRestrictMembers  bool `json:"can_restrict_members"`
	CanPromoteMembers   bool `json:"can_promote_members"`
	CanInviteUsers      bool `json:"can_invite_users"`
	CanManageVoiceChats bool `json:"can_manage_voice_chats"`
}

type PromoteChatMemberResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      bool   `json:"result"`
}

func PromoteChatMember(tgtoken string, chatid, userid int64) (bool, error) {
	// https://core.telegram.org/bots/api#promotechatmember

	req := PromoteChatMemberRequest{
		ChatId: chatid,
		UserId: userid,

		IsAnonymous:         false,
		CanManageChat:       true,
		CanPostMessages:     true,
		CanEditMessages:     true,
		CanDeleteMessages:   true,
		CanChangeInfo:       true,
		CanRestrictMembers:  true,
		CanPromoteMembers:   true,
		CanInviteUsers:      true,
		CanManageVoiceChats: true,
	}
	reqjson, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	var resp PromoteChatMemberResponse
	err = postJson(
		fmt.Sprintf("%s/bot%s/promoteChatMember", ApiUrlBase, tgtoken),
		bytes.NewBuffer(reqjson),
		&resp,
	)
	if err != nil {
		return false, fmt.Errorf("postJson: %w", err)
	}

	if !resp.Ok {
		return false, fmt.Errorf("%s", resp.Description)
	}

	return resp.Result, nil
}

type GetChatResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      Chat   `json:"result"`
}

func GetChat(tgtoken string, chatid int64) (chat Chat, err error) {
	// TODO too many requests retry

	requrl := fmt.Sprintf("%s/bot%s/getChat?chat_id=%d", ApiUrlBase, tgtoken, chatid)
	var resp GetChatResponse

	err = getJson(requrl, &resp, nil)
	if err != nil {
		return Chat{}, err
	}
	if !resp.Ok {
		return Chat{}, fmt.Errorf("telegram response not ok: %s", resp.Description)
	}

	return resp.Result, nil
}

type GetChatAdministratorsRequest struct {
	ChatId string `json:"chat_id"`
}

type ChatMember struct {
	User   User   `json:"user"`
	Status string `json:"status"`
}

type GetChatAdministratorsResponse struct {
	Ok          bool         `json:"ok"`
	Description string       `json:"description"`
	Result      []ChatMember `json:"result"`
}

func GetChatAdministrators(tgtoken string, chatid int64) (mm []ChatMember, err error) {
	requrl := fmt.Sprintf("%s/bot%s/getChatAdministrators?chat_id=%d", ApiUrlBase, tgtoken, chatid)
	var resp GetChatAdministratorsResponse

	err = getJson(requrl, &resp, nil)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("getChatAdministrators: %s", resp.Description)
	}

	return resp.Result, nil
}

type ChatMemberUpdated struct {
	Chat Chat  `json:"chat"`
	From User  `json:"from"`
	Date int64 `json:"date"`

	OldChatMember ChatMember `json:"old_chat_member"`
	NewChatMember ChatMember `json:"new_chat_member"`

	ViaJoinRequest          bool `json:"via_join_request"`
	ViaChatFolderInviteLink bool `json:"via_chat_folder_invite_link"`
}

type Update struct {
	UpdateId int64 `json:"update_id"`

	Message           Message `json:"message"`
	EditedMessage     Message `json:"edited_message"`
	ChannelPost       Message `json:"channel_post"`
	EditedChannelPost Message `json:"edited_channel_post"`

	MyChatMemberUpdated ChatMemberUpdated `json:"my_chat_member"`
}

type GetUpdatesResponse struct {
	Ok          bool     `json:"ok"`
	Description string   `json:"description"`
	Result      []Update `json:"result"`
}

func GetUpdates(tgtoken string, offset int64) (uu []Update, respjson string, err error) {
	requrl := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d", ApiUrlBase, tgtoken, offset)

	var resp GetUpdatesResponse
	err = getJson(requrl, &resp, &respjson)
	if err != nil {
		return nil, "", err
	}
	if !resp.Ok {
		return nil, "", fmt.Errorf("telegram response not ok: %s", resp.Description)
	}

	return resp.Result, respjson, nil
}

func getJson(requrl string, result interface{}, respjson *string) (err error) {
	resp, err := HttpClient.Get(requrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	err = json.NewDecoder(bytes.NewBuffer(respBody)).Decode(result)
	if err != nil {
		return fmt.Errorf("json.Decode: %w", err)
	}

	if DEBUG {
		log("DEBUG getJson %s response ContentLength==%d Body==```"+NL+"%s"+NL+"```", requrl, resp.ContentLength, respBody)
	}
	if respjson != nil {
		*respjson = string(respBody)
	}

	return nil
}

func postJson(requrl string, data *bytes.Buffer, result interface{}) error {
	resp, err := HttpClient.Post(
		requrl,
		"application/json",
		data,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody := bytes.NewBuffer(nil)
	_, err = io.Copy(respBody, resp.Body)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	err = json.NewDecoder(respBody).Decode(result)
	if err != nil {
		return fmt.Errorf("json.Decode: %v", err)
	}

	if DEBUG {
		log("DEBUG postJson %s response ContentLength==%d Body==```"+NL+"%s"+NL+"```", requrl, resp.ContentLength, respBody)
	}

	return nil
}

func ts() string {
	tnow := time.Now().UTC()
	return fmt.Sprintf(
		"%d%:02d%02d:%02d%02d+",
		tnow.Year()%1000, tnow.Month(), tnow.Day(),
		tnow.Hour(), tnow.Minute(),
	)
}

func log(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, ts()+" "+msg+NL, args...)
}
