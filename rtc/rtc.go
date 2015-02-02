package rtc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/fcoury/rtc-go/browser"
	"github.com/fcoury/rtc-go/models"
)

type RTC struct {
	User     string
	Password string

	browser *browser.Browser
}

type WorkItem struct {
	Id           string
	Type         string
	Summary      string
	CreatedBy    string
	OwnedBy      string
	Estimate     string
	FiledAgainst string
	PlannedFor   string
	LocationUri  string
}

func NewRTC(user string, password string) *RTC {
	return &RTC{User: user, Password: password, browser: browser.NewBrowser(false)}
}

func (rtc *RTC) request(method string, url string, data string) (*http.Response, error) {
	return rtc.browser.Request(method, url, data)
}

func (rtc *RTC) requestBody(method string, url string, data string) ([]byte, error) {
	resp, err := rtc.browser.Request(method, url, data)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, err
}

func (rtc *RTC) Login() error {
	_, err := rtc.request("GET", "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/j_security_check?j_username=fcoury%40br.ibm.com&j_password=tempra14", "")

	if err != nil {
		return err
	}

	_, err = rtc.request("POST", "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/", "")

	if err != nil {
		return err
	}

	return nil
}

func (rtc *RTC) CurrentWorkItems() ([]*WorkItem, error) {
	var workItems []*WorkItem

	data := "startIndex=0&maxResults=100&absoluteURIs=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&columnIdentifiers=workItemType&columnIdentifiers=summary&columnIdentifiers=creator&columnIdentifiers=owner&columnIdentifiers=creationDate&columnIdentifiers=duration&columnIdentifiers=category&columnIdentifiers=target&columnIdentifiers=projectArea&columnIdentifiers=internalTags&itemId=_VMvycVRcEd61fuNW84kdiQ&skipOAuth=true&filterAttribute=&filterValue="
	queryUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet"

	resp, err := rtc.request("POST", queryUrl, data)
	if err != nil {
		return workItems, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return workItems, err
	}

	env, err := models.NewFromXml(body)

	if err != nil {
		return workItems, err
	}

	for _, row := range env.Body.Response.ReturnValue.Value.Rows {
		// fmt.Printf("Row: %+v\n", row)
		// for i, l := range row.Labels {
		// 	fmt.Printf("%d - %s\n", i, l)
		// }
		// fmt.Printf("Item: %d - %s\n", row.Id, row.Labels[1])
		// 0 - Task
		// 1 - Analysis: Component elimination. Removal of order component failed.
		// 2 - Marcelo De Campos
		// 3 - Felipe Gonçalves Coury
		// 4 - 1374758469269
		// 5 - 24 hours
		// 6 - Unassigned
		// 7 - [2014] February R1, S1
		// 8 - SD-OPS
		// 9 -
		wi := &WorkItem{
			Id:           row.Id,
			Type:         row.Labels[0],
			Summary:      row.Labels[1],
			CreatedBy:    row.Labels[2],
			OwnedBy:      row.Labels[3],
			Estimate:     row.Labels[5],
			FiledAgainst: row.Labels[6],
			PlannedFor:   row.Labels[7],
			LocationUri:  row.LocationUri,
		}
		workItems = append(workItems, wi)
	}

	return workItems, err
}

func (rtc *RTC) Search(query string) ([]*WorkItem, error) {
	var workItems []*WorkItem

	url := fmt.Sprintf("https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/results?maxResults=100&fullText=%s&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", url.QueryEscape(query))

	body, err := rtc.requestBody("GET", url, "")
	if err != nil {
		return workItems, err
	}

	env, err := models.NewFromXml(body)
	if err != nil {
		return workItems, err
	}

	for _, twi := range env.Body.Response.ReturnValue.Value.WorkItems {
		wi := &WorkItem{
			Id:          twi.Id,
			Summary:     twi.Summary,
			Type:        twi.Type,
			OwnedBy:     twi.OwnerName,
			LocationUri: twi.LocationUri,
		}
		workItems = append(workItems, wi)
	}

	return workItems, nil
}
