package models

import (
	"encoding/xml"
	"regexp"
	"strings"
)

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
	Type       string      `xml:"type"`
	Value      Value       `xml:"value"`
	ItemValues []ItemValue `xml:"UIItemSetDTO"`
	Releases   []Release   `xml:"IterationDTO"`
}

func (r ReturnValue) GetItems() map[string]map[string]string {
	var m map[string]map[string]string
	m = make(map[string]map[string]string)

	for _, iv := range r.ItemValues {
		m[iv.AttributeName] = iv.GetPairs()
	}

	return m
}

type ItemValue struct {
	AttributeName string `xml:"attributeName"`
	Items         []Item `xml:"uiItems"`
}

func (i ItemValue) GetPairs() map[string]string {
	var m map[string]string
	m = make(map[string]string)

	for _, it := range i.Items {
		m[it.Id] = it.Label
	}

	return m
}

type Item struct {
	Id          string `xml:"id"`
	Label       string `xml:"label"`
	ItemType    string `xml:"itemType"`
	IsNullValue string `xml:"isNullValue"`
	Arguments   string `xml:"arguments"`
	IsArchived  string `xml:"isArchived"`
	Current     string `xml:"current"`
	InScope     string `xml:"inScope"`
	StartDate   string `xml:"startDate"`
	EndDate     string `xml:"endDate"`
}

type Release struct {
	Id         string      `xml:"id"`
	ItemId     string      `xml:"itemid"`
	Label      string      `xml:"label"`
	Iterations []Iteration `xml:"iterations"`
}

type Iteration struct {
	Id        string `xml:"id"`
	ItemId    string `xml:"itemId"`
	Label     string `xml:"label"`
	Completed string `xml:"completed"`
	Archived  string `xml:"archived"`
}

type Value struct {
	StartIndex     int64        `xml:"startIndex"`
	TotalCount     int64        `xml:"totalCount"`
	EstimatedTotal int64        `xml:"estimatedTotal"`
	Limit          int64        `xml:"limit"`
	StateId        string       `xml:"stateId"`
	ItemId         string       `xml:"itemId"`
	Token          string       `xml:"resultToken"`
	Headers        []Header     `xml:"headers"`
	Rows           []Row        `xml:"rows"`
	WorkItems      []WorkItem   `xml:"workItemSummaryDTOs"`
	WorkItem       WorkItem     `xml:"workItem"`
	Attributes     []*Attribute `xml:"attributes"`
	LinkTypes      []LinkType   `xml:"linkTypes"`
}

func (val Value) GetAttributes() []*Attribute {
	return val.Attributes
}

type LinkType struct {
	Id               string `xml:"id"`
	DisplayName      string `xml:"displayName"`
	IconUrl          string `xml:"iconUrl"`
	IsSingleValued   string `xml:"isSingleValued"`
	IsSource         string `xml:"isSource"`
	IsSymmetric      string `xml:"isSymmetric"`
	ItemType         string `xml:"itemType"`
	IsUserDeleteable string `xml:"isUserDeleteable"`
	IsUserWriteable  string `xml:"isUserWriteable"`
	EndpointId       string `xml:"endpointId"`
	IsInternal       string `xml:"isInternal"`
	Links            []Link `xml:"linkDTOs"`
}

type Link struct {
	ItemId      string     `xml:"itemId"`
	StateId     string     `xml:"stateId"`
	locationUri string     `xml:"weburi"`
	Target      LinkTarget `xml:"target"`
}

type LinkTarget struct {
	ItemId      string       `xml:"itemId"`
	StateId     string       `xml:"stateId"`
	LocationUri string       `xml:"locationUri"`
	Attributes  []*Attribute `xml:"attributes"`
}

func (link LinkTarget) GetAttributes() []*Attribute {
	return link.Attributes
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

type WorkItem struct {
	WorkItemId  string       `xml:"workItemItemId"`
	Id          string       `xml:"id"`
	ItemId      string       `xml:"itemId"`
	StateId     string       `xml:"stateId"`
	Summary     string       `xml:"summary"`
	OwnerName   string       `xml:"ownerName"`
	CreatorName string       `xml:"creatorName"`
	Type        string       `xml:"typeName"`
	LocationUri string       `xml:"locationUri"`
	Description string       `xml:"description"`
	Attributes  []*Attribute `xml:"attributes"`
}

func (wi WorkItem) GetAttributes() []*Attribute {
	return wi.Attributes
}

type Attribute struct {
	Key   string     `xml:"key"`
	Value *AttrValue `xml:"value"`
}

type AttrValue struct {
	Label   string `xml:"label"`
	Content string `xml:"content"`
	Id      string `xml:"id"`
}

func NewFromXml(xmlData []byte) (*Envelope, error) {
	var b *Envelope

	xlated := string(xmlData)
	xlated = strings.Replace(xlated, "\n", "", -1)

	reg, err := regexp.Compile(`<values xsi:type="(.*?):(.*?)">(.*?)</values>`)
	if err != nil {
		return nil, err
	}

	safe := reg.ReplaceAllString(xlated, "<$2>$3</$2>")

	// fmt.Println(safe)

	err = xml.Unmarshal([]byte(safe), &b)

	return b, err
}

func (wi *WorkItem) ParsedId() string {
	s := strings.Split(wi.LocationUri, "/")
	return s[len(s)-1]
}

func (v *Value) Attribute(key string) *AttrValue {
	for _, a := range v.Attributes {
		if a.Key == key {
			return a.Value
		}
	}

	return nil
}
