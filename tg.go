/*
history:
025/0404 v1

https://core.telegram.org/bots/api

GoFmt GoBuildNull
*/

package tg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	NL = "\n"

	// https://core.telegram.org/bots/api#formatting-options
	ParseMode = "MarkdownV2"
)

var (
	HttpClient = &http.Client{}

	DEBUG = false

	ApiUrl   = "https://api.telegram.org"
	ApiToken = ""
)

func Esc(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	// https://core.telegram.org/bots/api#formatting-options
	for _, c := range "\\_*[]()~`>#+-=|{}.!" {
		text = strings.ReplaceAll(text, string(c), "\\"+string(c))
	}
	return text
}

// escape with exceptions
func EscExcept(text string, except string) string {
	// https://core.telegram.org/bots/api#formatting-options
	// https://pkg.go.dev/strings#ReplaceAll
	escape := "\\[]()~`>#+-=|{}.!"
	for _, c := range except {
		escape = strings.ReplaceAll(escape, string(c), "")
	}
	for _, c := range escape {
		text = strings.ReplaceAll(text, string(c), "\\"+string(c))
	}
	return text
}

func Bold(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	// https://core.telegram.org/bots/api#formatting-options
	return "*" + text + "*"
}

func Italic(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	// https://core.telegram.org/bots/api#formatting-options
	return "_" + text + "_"
}

func Underline(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	// https://core.telegram.org/bots/api#formatting-options
	return "__" + text + "__"
}

func BoldUnderline(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	// https://core.telegram.org/bots/api#formatting-options
	return "__*" + text + "*__"
}

func ItalicUnderline(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	// https://core.telegram.org/bots/api#formatting-options
	return "_ __" + text + "__ _"
}

func Spoiler(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	// https://core.telegram.org/bots/api#formatting-options
	return "||" + text + "||"
}

func Code(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	for _, c := range "\\`" {
		text = strings.ReplaceAll(text, string(c), "\\"+string(c))
	}
	return "`" + text + "`"
}

func Link(text, url string) string {
	for _, c := range "\\)" {
		url = strings.ReplaceAll(url, string(c), "\\"+string(c))
	}
	return fmt.Sprintf("[%s](%s)", Esc(text), url)
}

func Pre(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	for _, c := range "\\`" {
		text = strings.ReplaceAll(text, string(c), "\\"+string(c))
	}
	return "```" + NL + text + NL + "```"
}

func Quote(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	text = ">" + text
	text = strings.ReplaceAll(text, NL, NL+">")
	return text + NL
}

func ExpandQuote(text string, args ...interface{}) string {
	text = fmt.Sprintf(text, args...)
	text = ">" + text
	text = strings.ReplaceAll(text, NL, NL+">")
	text += "||"
	return text + NL
}

type Message struct {
	// https://core.telegram.org/bots/api#message
	Id        string
	MessageId int64  `json:"message_id"`
	From      User   `json:"from,omitempty"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`

	ReplyToMessage struct {
		MessageId  int64  `json:"message_id"`
		From       User   `json:"from"`
		SenderChat Chat   `json:"sender_chat"`
		Chat       Chat   `json:"chat"`
		Date       int64  `json:"date"`
		Text       string `json:"text"`
	} `json:"reply_to_message"`

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

type MessageResponse struct {
	Ok          bool     `json:"ok"`
	Description string   `json:"description"`
	Result      *Message `json:"result"`
}

func SendMessage(req SendMessageRequest) (msg *Message, err error) {
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

	requrl := fmt.Sprintf("%s/bot%s/sendMessage", ApiUrl, ApiToken)
	var resp MessageResponse

	if err := postJson(requrl, bytes.NewBuffer(reqjson), &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("sendMessage: %s", resp.Description)
	}

	msg = resp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	return msg, nil
}

type ReactionTypeEmoji struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type SetMessageReactionRequest struct {
	// https://core.telegram.org/bots/api#setmessagereaction

	ChatId    string `json:"chat_id"`
	MessageId int64  `json:"message_id"`

	Reaction []ReactionTypeEmoji `json:"reaction"`
}

func SetMessageReaction(req SetMessageReactionRequest) (err error) {
	// https://core.telegram.org/bots/api#setmessagereaction

	if DEBUG {
		log("DEBUG req==%#v", req)
	}
	for i, _ := range req.Reaction {
		req.Reaction[i].Type = "emoji"
	}
	reqjson, err := json.Marshal(req)
	if err != nil {
		return err
	}

	requrl := fmt.Sprintf("%s/bot%s/setMessageReaction", ApiUrl, ApiToken)
	var resp BoolResponse

	if err := postJson(requrl, bytes.NewBuffer(reqjson), &resp); err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("setMessageReaction: %s", resp.Description)
	}
	if !resp.Result {
		return fmt.Errorf("setMessageReaction: %s", resp.Description)
	}

	return nil
}

type PhotoSize struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
	Width        int64  `json:"width"`
	Height       int64  `json:"height"`
	FileSize     int64  `json:"file_size"`
}

type SendPhotoFileRequest struct {
	ChatId   string
	FileName string
	Photo    io.Reader
}

// https://core.telegram.org/bots/api#sendphoto
func SendPhotoFile(req SendPhotoFileRequest) (msg *Message, err error) {
	var mpartBuf bytes.Buffer
	mpart := multipart.NewWriter(&mpartBuf)
	var formWr io.Writer

	// chat_id
	err = mpart.WriteField("chat_id", req.ChatId)
	if err != nil {
		return nil, fmt.Errorf("WriteField chat_id: %v", err)
	}

	// photo
	formWr, err = mpart.CreateFormFile("photo", req.FileName)
	if err != nil {
		return nil, fmt.Errorf("CreateFormFile photo: %v", err)
	}
	_, err = io.Copy(formWr, req.Photo)
	if err != nil {
		return nil, fmt.Errorf("Copy photo: %v", err)
	}

	err = mpart.Close()
	if err != nil {
		return nil, fmt.Errorf("multipartWriter.Close: %v", err)
	}

	resp, err := HttpClient.Post(
		fmt.Sprintf("%s/bot%s/sendPhoto", ApiUrl, ApiToken),
		mpart.FormDataContentType(),
		&mpartBuf,
	)
	if err != nil {
		return nil, fmt.Errorf("Post: %v", err)
	}
	defer resp.Body.Close()

	var tgresp MessageResponse
	err = json.NewDecoder(resp.Body).Decode(&tgresp)
	if err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}
	if !tgresp.Ok {
		return nil, fmt.Errorf("sendPhoto: %s", tgresp.Description)
	}

	msg = tgresp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	if len(msg.Photo) == 0 {
		return nil, fmt.Errorf("sendPhoto: Photo array empty")
	}

	return msg, nil
}

type SendPhotoRequest struct {
	ChatId    string `json:"chat_id"`
	Photo     string `json:"photo"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

func SendPhoto(req SendPhotoRequest) (msg *Message, err error) {
	// https://core.telegram.org/bots/api#sendphoto

	if DEBUG {
		log("DEBUG SendPhoto req==%#v", req)
	}
	if req.ParseMode == "" {
		req.ParseMode = ParseMode
	}
	reqjson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	requrl := fmt.Sprintf("%s/bot%s/sendPhoto", ApiUrl, ApiToken)

	var resp MessageResponse
	if err := postJson(requrl, bytes.NewBuffer(reqjson), &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("sendPhoto: %s", resp.Description)
	}

	msg = resp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	return msg, nil
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

type SendAudioFileRequest struct {
	ChatId    string
	Caption   string
	Performer string
	Title     string
	Duration  time.Duration
	Audio     io.Reader
	Thumb     io.Reader
}

func SendAudioFile(req SendAudioFileRequest) (msg *Message, err error) {
	// https://core.telegram.org/bots/api#sending-files

	if req.Audio == nil {
		return nil, fmt.Errorf("Audio is nil")
	}

	var mpartBuf bytes.Buffer
	mpart := multipart.NewWriter(&mpartBuf)

	// chat_id
	if err := mpart.WriteField("chat_id", req.ChatId); err != nil {
		return nil, fmt.Errorf("WriteField chat_id: %v", err)
	}

	// caption
	if err := mpart.WriteField("caption", req.Caption); err != nil {
		return nil, fmt.Errorf("WriteField caption: %v", err)
	}

	// performer
	if err := mpart.WriteField("performer", req.Performer); err != nil {
		return nil, fmt.Errorf("WriteField performer: %v", err)
	}

	// title
	if err := mpart.WriteField("title", req.Title); err != nil {
		return nil, fmt.Errorf("WriteField title: %v", err)
	}

	// duration
	if err := mpart.WriteField("duration", strconv.Itoa(int(req.Duration.Seconds()))); err != nil {
		return nil, fmt.Errorf("WriteField duration: %v", err)
	}

	filename := safestring(req.Performer + "." + req.Title)

	// audio
	if w, err := mpart.CreateFormFile("audio", filename); err != nil {
		return nil, fmt.Errorf("CreateFormFile audio: %v", err)
	} else if _, err := io.Copy(w, req.Audio); err != nil {
		return nil, fmt.Errorf("Copy audio: %v", err)
	}

	if req.Thumb != nil {
		// thumb
		if w, err := mpart.CreateFormFile("thumb", filename); err != nil {
			return nil, fmt.Errorf("CreateFormFile thumb: %v", err)
		} else if _, err := io.Copy(w, req.Thumb); err != nil {
			return nil, fmt.Errorf("Copy thumb: %v", err)
		}
	}

	if err := mpart.Close(); err != nil {
		return nil, fmt.Errorf("multipart.Writer.Close: %v", err)
	}

	resp, err := HttpClient.Post(
		fmt.Sprintf("%s/bot%s/sendAudio", ApiUrl, ApiToken),
		mpart.FormDataContentType(),
		&mpartBuf,
	)
	if err != nil {
		return nil, fmt.Errorf("Post: %v", err)
	}
	defer resp.Body.Close()

	var tgresp MessageResponse
	err = json.NewDecoder(resp.Body).Decode(&tgresp)
	if err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}
	if !tgresp.Ok {
		return nil, fmt.Errorf("sendAudio: %s", tgresp.Description)
	}

	msg = tgresp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	if msg.Audio.FileId == "" {
		return nil, fmt.Errorf("sendAudio: Audio.FileId empty")
	}

	return msg, nil
}

type SendAudioRequest struct {
	ChatId    string `json:"chat_id"`
	Audio     string `json:"audio"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

func SendAudio(req SendAudioRequest) (msg *Message, err error) {
	// https://core.telegram.org/bots/API#sendaudio

	if req.ParseMode == "" {
		req.ParseMode = ParseMode
	}

	reqjson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	requrl := fmt.Sprintf("%s/bot%s/sendAudio", ApiUrl, ApiToken)

	var resp MessageResponse
	if err := postJson(requrl, bytes.NewBuffer(reqjson), &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("sendAudio: %s", resp.Description)
	}

	msg = resp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	return msg, nil
}

type SendVideoFileRequest struct {
	ChatId        string
	Caption       string
	Video         io.Reader
	Width, Height int
	Duration      time.Duration
}

func SendVideoFile(req SendVideoFileRequest) (msg *Message, err error) {
	if req.Video == nil {
		return nil, fmt.Errorf("Video is nil")
	}

	piper, pipew := io.Pipe()
	mpartw := multipart.NewWriter(pipew)

	var mparterr error
	go func(err error) {
		defer func() {
			if mparterr != nil {
				log("mparterr: %v", err)
			}
		}()

		var formw io.Writer

		defer pipew.Close()

		// chat_id
		err = mpartw.WriteField("chat_id", req.ChatId)
		if err != nil {
			err = fmt.Errorf("WriteField chat_id: %w", err)
			return
		}

		// caption
		err = mpartw.WriteField("caption", req.Caption)
		if err != nil {
			err = fmt.Errorf("WriteField caption: %w", err)
			return
		}

		// width
		err = mpartw.WriteField("width", strconv.Itoa(req.Width))
		if err != nil {
			err = fmt.Errorf("WriteField width: %w", err)
			return
		}

		// height
		err = mpartw.WriteField("height", strconv.Itoa(req.Height))
		if err != nil {
			err = fmt.Errorf("WriteField height: %w", err)
			return
		}

		// duration
		err = mpartw.WriteField("duration", strconv.Itoa(int(req.Duration.Seconds())))
		if err != nil {
			err = fmt.Errorf("CreateFormField(`duration`): %w", err)
			return
		}

		filename := safestring(req.Caption)

		// video
		formw, err = mpartw.CreateFormFile("video", filename)
		if err != nil {
			err = fmt.Errorf("CreateFormFile('video'): %w", err)
			return
		}
		_, err = io.Copy(formw, req.Video)
		if err != nil {
			err = fmt.Errorf("Copy req.Video: %w", err)
			return
		}

		if err := mpartw.Close(); err != nil {
			err = fmt.Errorf("multipart.Writer.Close: %w", err)
			return
		}
	}(mparterr)

	resp, err := HttpClient.Post(
		fmt.Sprintf("%s/bot%s/sendVideo", ApiUrl, ApiToken),
		mpartw.FormDataContentType(),
		piper,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if mparterr != nil {
		return nil, err
	}

	var tgresp MessageResponse
	err = json.NewDecoder(resp.Body).Decode(&tgresp)
	if err != nil {
		return nil, fmt.Errorf("Decode: %w", err)
	}
	if !tgresp.Ok {
		return nil, fmt.Errorf("sendVideo: %s", tgresp.Description)
	}

	msg = tgresp.Result
	msg.Id = fmt.Sprintf("%d", msg.MessageId)

	if msg.Video.FileId == "" {
		return nil, fmt.Errorf("sendVideo: Video.FileId empty")
	}

	return msg, nil
}

type DeleteMessageRequest struct {
	ChatId    string `json:"chat_id"`
	MessageId int64  `json:"message_id"`
}

type BoolResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      bool   `json:"result"`
}

func DeleteMessage(req DeleteMessageRequest) error {
	// https://core.telegram.org/bots/api#deletemessage

	reqjson, err := json.Marshal(req)
	if err != nil {
		return err
	}

	requrl := fmt.Sprintf("%s/bot%s/deleteMessage", ApiUrl, ApiToken)
	var resp BoolResponse

	if err := postJson(requrl, bytes.NewBuffer(reqjson), &resp); err != nil {
		return fmt.Errorf("postJson: %w", err)
	}
	if !resp.Ok {
		return fmt.Errorf("deleteMessage: %s", resp.Description)
	}

	return nil
}

type PromoteChatMemberRequest struct {
	ChatId string `json:"chat_id"`
	UserId string `json:"user_id"`

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

func PromoteChatMember(chatid, userid string) (bool, error) {
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

	requrl := fmt.Sprintf("%s/bot%s/promoteChatMember", ApiUrl, ApiToken)
	var resp BoolResponse

	if err := postJson(requrl, bytes.NewBuffer(reqjson), &resp); err != nil {
		return false, fmt.Errorf("postJson: %w", err)
	}
	if !resp.Ok {
		return false, fmt.Errorf("%s", resp.Description)
	}

	return resp.Result, nil
}

type ChatResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      Chat   `json:"result"`
}

func GetChat(chatid int64) (chat Chat, err error) {
	// TODO too many requests retry

	requrl := fmt.Sprintf("%s/bot%s/getChat?chat_id=%d", ApiUrl, ApiToken, chatid)
	var resp ChatResponse

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

type ChatMembersResponse struct {
	Ok          bool         `json:"ok"`
	Description string       `json:"description"`
	Result      []ChatMember `json:"result"`
}

func GetChatAdministrators(chatid int64) (mm []ChatMember, err error) {
	requrl := fmt.Sprintf("%s/bot%s/getChatAdministrators?chat_id=%d", ApiUrl, ApiToken, chatid)
	var resp ChatMembersResponse

	if err := getJson(requrl, &resp, nil); err != nil {
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

// https://core.telegram.org/bots/api#update
type Update struct {
	UpdateId int64 `json:"update_id"`

	Message       Message `json:"message"`
	EditedMessage Message `json:"edited_message"`

	ChannelPost       Message `json:"channel_post"`
	EditedChannelPost Message `json:"edited_channel_post"`

	MyChatMember ChatMemberUpdated `json:"my_chat_member"`
	ChatMember   ChatMemberUpdated `json:"chat_member"`
}

type UpdatesResponse struct {
	Ok          bool     `json:"ok"`
	Description string   `json:"description"`
	Result      []Update `json:"result"`
}

func GetUpdates(offset int64) (uu []Update, respjson string, err error) {
	requrl := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d", ApiUrl, ApiToken, offset)

	var resp UpdatesResponse
	err = getJson(requrl, &resp, &respjson)
	if err != nil {
		return nil, "", err
	}
	if !resp.Ok {
		return nil, "", fmt.Errorf("telegram response not ok: %s", resp.Description)
	}

	return resp.Result, respjson, nil
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

func safestring(s string) (t string) {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			r = '.'
		}
		t = t + string(r)
	}

	if len([]rune(t)) > 40 {
		t = string([]rune(t)[:40])
	}

	return t
}

func ts() string {
	tnow := time.Now().In(time.FixedZone("IST", 330*60))
	return fmt.Sprintf(
		"%d%02d%02d:%02d%02d+",
		tnow.Year()%1000, tnow.Month(), tnow.Day(),
		tnow.Hour(), tnow.Minute(),
	)
}

func log(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, ts()+" "+msg+NL, args...)
}
