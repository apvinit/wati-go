package wati

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	baseUrl string
	token   string
	http    *http.Client
}

func NewClient(baseUrl, token string) *Client {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	return &Client{
		baseUrl: baseUrl,
		token:   token,
		http:    client,
	}
}

func (c *Client) do(method, endpoint string, body io.Reader, params map[string]string) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s%s", c.baseUrl, endpoint)
	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.token)
	q := req.URL.Query()
	for key, val := range params {
		q.Set(key, val)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := c.http.Do(req)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	return b, err
}
func (c *Client) doMultipart(method, endpoint string, body io.Reader, params map[string]string) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s%s", c.baseUrl, endpoint)
	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Add("Authorization", "Bearer "+c.token)
	q := req.URL.Query()
	for key, val := range params {
		q.Set(key, val)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := c.http.Do(req)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	return b, err
}

func (c *Client) GetMessages(whatsappNumber string) ([]byte, error) {
	b, err := c.do("GET", "/api/v1/getMessages/"+whatsappNumber, nil, nil)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) GetMessageTemplates() ([]byte, error) {
	b, err := c.do("GET", "/api/v1/getMessageTemplates", nil, nil)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) GetContactsList() ([]byte, error) {
	b, err := c.do("GET", "/api/v1/getContacts", nil, nil)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// TODO : figure out what is the filename and where to find it
func (c *Client) GetMedia(fileName string) ([]byte, error) {
	b, err := c.do("GET", "/api/v1/getMedia", nil, map[string]string{
		"fileName": fileName,
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}

type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) UpdateContactAttributes(whatsappNumber string, params []Param) ([]byte, error) {
	body, err := json.Marshal(struct {
		CustomParams []Param `json:"customParams"`
	}{params})
	if err != nil {
		return nil, err
	}
	return c.do("POST", "/api/v1/updateContactAttributes/"+whatsappNumber, bytes.NewBuffer(body), nil)
}

func (c *Client) RotateToken(token string) ([]byte, error) {
	return c.do("POST", "/api/v1/rotateToken", nil, map[string]string{
		"token": token,
	})
}

func (c *Client) AddContact(whatsappNumber, name string, params []Param) ([]byte, error) {

	data := struct {
		Name         string  `json:"name"`
		CustomParams []Param `json:"customParams"`
	}{
		name, params,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	res, err := c.do("POST", "/api/v1/addContact/"+whatsappNumber, bytes.NewBuffer(body), nil)
	return res, err
}

func (c *Client) SendSessionFile(whatsappNumber string, caption string, file *os.File) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	return c.doMultipart("POST", "/api/v1/sendSessionFile/"+whatsappNumber, body, map[string]string{
		"caption": caption,
	})
}

func (c *Client) SendSessionMessage(whatsappNumber, messageText string) ([]byte, error) {

	return c.do("POST", "/api/v1/sendSessionMessage/"+whatsappNumber, nil, map[string]string{
		"messageText": messageText,
	})

}

func (c *Client) SendTemplateMessage(whatsappNumber, templateName, broadcastName string, params []Param) ([]byte, error) {

	data := struct {
		TemplateName  string  `json:"template_name"`
		BroadcastName string  `json:"broadcast_name"`
		Parameters    []Param `json:"parameters"`
	}{
		TemplateName:  templateName,
		BroadcastName: broadcastName,
		Parameters:    params,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.do("POST", "/api/v1/sendTemplateMessage", bytes.NewBuffer(body), map[string]string{
		"whatsappNumber": whatsappNumber,
	})
}

type Receiver struct {
	WhatsappNumber string  `json:"whatsappNumber"`
	CustomParams   []Param `json:"customParams"`
}

func (c *Client) SendTemplateMessages(templateName, brodcastName string, receivers []Receiver) ([]byte, error) {
	data := struct {
		TemplateName  string     `json:"template_name"`
		BroadcastName string     `json:"broadcast_name"`
		Receivers     []Receiver `json:"receivers"`
	}{
		TemplateName:  templateName,
		BroadcastName: brodcastName,
		Receivers:     receivers,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.do("POST", "/api/v1/sendTemplateMessages", bytes.NewBuffer(body), nil)
}

func (c *Client) SendTemplateMessageCSV(templateName, broadcastName string) ([]byte, error) {
	data := struct {
		WhatsappNumbersCSV string `json:"whatsapp_numbers_csv"`
	}{}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.do("POST", "/api/v1/sendTemplateMessageCSV", bytes.NewBuffer(body), map[string]string{
		"template_name":  templateName,
		"broadcast_name": broadcastName,
	})
}

type Media struct {
	Url      string `json:"url"`
	FileName string `json:"fileName"`
}

type Header struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Media Media  `json:"media"`
}

type Btn struct {
	Text string `json:"text"`
}

func (c *Client) SendInteractiveButtonsMessage(whatsappNumber string, header Header, body, footer string, buttons []Btn) ([]byte, error) {
	data := struct {
		Header  Header `json:"header"`
		Body    string `json:"body"`
		Footer  string `json:"footer"`
		Buttons []Btn  `json:"buttons"`
	}{
		Header:  header,
		Body:    body,
		Footer:  footer,
		Buttons: buttons,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.do("POST", "/api/v1/sendInteractiveButtonsMessage", bytes.NewBuffer(b), map[string]string{
		"whatsappNumber": whatsappNumber,
	})
}

type Row struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Section struct {
	Title string `json:"title"`
	Rows  Row    `json:"rows"`
}

func (c *Client) SendInteractiveListMessage(whatsappNumber, header, body, footer, buttonText string, sections []Section) ([]byte, error) {
	data := struct {
		Header     string    `json:"header"`
		Body       string    `json:"body"`
		Footer     string    `json:"footer"`
		ButtonText string    `json:"buttonText"`
		Sections   []Section `json:"sections"`
	}{
		Header:     header,
		Body:       body,
		Footer:     footer,
		ButtonText: buttonText,
		Sections:   sections,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.do("POST", "/api/v1/sendInteractiveListMessage", bytes.NewBuffer(b), map[string]string{
		"whatsappNumber": whatsappNumber,
	})

}

func (c *Client) AssignOperator(email, whatsappNumber string) ([]byte, error) {
	b, err := c.do("GET", "/api/v1/getMedia", nil, map[string]string{
		"email":          email,
		"whatsappNumber": whatsappNumber,
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}
