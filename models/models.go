package models

import "encoding/xml"

type Envelope struct {
	Body Body `xml:"Body"`
}

type Body struct {
	Response Response `xml:"response"`
}

type Response struct {
	Method      string      `xml:"method"`
	Interface   string      `xml:"interface"`
	ReturnValue ReturnValue `xml:"returnValue"`
}

type ReturnValue struct {
	Type  string `xml:"type"`
	Value Value  `xml:"value"`
}

type Value struct {
	StartIndex     int64 `xml:"startIndex"`
	TotalCount     int64 `xml:"totalCount"`
	EstimatedTotal int64 `xml:"estimatedTotal"`
	Limit          int64 `xml:"limit"`
	Token          string
	Headers        []Header `xml:"headers"`
	Rows           []Row    `xml:"rows"`
}

type Header struct {
	AttributeId   string `xml:"attributeId"`
	AttributeType string `xml:"attributeType"`
	Label         string `xml:"label"`
	IsOrderable   string `xml:"isOrderable"`
	Alignment     string `xml:"alignment"`
	Width         string `xml:"width"`
	IconOnly      string `xml:"iconOnly"`
	LabelOnly     string `xml:"labelOnly"`
}

type Row struct {
	Id          string   `xml:"id"`
	ItemId      string   `xml:"itemId"`
	StateGroup  string   `xml:"stateGroup"`
	Labels      []string `xml:"labels"`
	LocationUri string   `xml:"locationUri"`
}

func NewFromXml(xmlData []byte) (Envelope, error) {
	var b Envelope
	err := xml.Unmarshal(xmlData, &b)

	return b, err
}
