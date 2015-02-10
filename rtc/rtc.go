package rtc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/fcoury/rtc-go/browser"
	"github.com/fcoury/rtc-go/models"
	"github.com/skratchdot/open-golang/open"
)

type RTC struct {
	User     string
	Password string

	OwnerId string

	browser *browser.Browser
}

type WorkItem struct {
	Id           string
	ItemId       string
	StateId      string
	Type         string
	Summary      string
	CreatedBy    string
	OwnedBy      string
	Estimate     string
	TimeSpent    string
	FiledAgainst string
	PlannedFor   string
	LocationUri  string
	Description  string
	State        string
	Resolution   string
	IterationId  string
	CodeChanges  string
	Iteration    Iteration
	Parents      []Reference
	Children     []Reference
	Approvals    []models.Approval
}

type Reference struct {
	ItemId      string
	StateId     string
	LocationUri string
	Summary     string
	Type        string
	Id          string
}

type Release struct {
	Id         string
	ItemId     string
	Label      string
	Completed  string
	Archived   string
	Iterations []Iteration
}

type Iteration struct {
	Id        string
	ItemId    string
	Label     string
	Completed string
	Archived  string
}

type Owner struct {
	Id   string
	Name string
}

type byName []Owner

func (o byName) Len() int {
	return len(o)
}
func (o byName) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
func (o byName) Less(i, j int) bool {
	return o[i].Name < o[j].Name
}

func (wi *WorkItem) Title() string {
	return fmt.Sprintf("%s %s - %s", wi.Type, wi.Id, wi.Summary)
}

func (wi *WorkItem) Owner() string {
	parts := strings.Split(wi.OwnedBy, " ")
	res := ""
	for _, c := range parts {
		res = res + string(c[0])
	}
	return res
}

func NewRTC(user string, password string, ownerId string) *RTC {
	return &RTC{User: user, Password: password, OwnerId: ownerId, browser: browser.NewBrowser(false)}
}

func (rtc *RTC) request(method string, url string, data string) (*http.Response, error) {
	return rtc.browser.Request(method, url, data)
}

func (rtc *RTC) requestBody(method string, url string, data string) ([]byte, error) {
	resp, err := rtc.browser.Request(method, url, data)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("Error: got HTTP status code %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	// fmt.Println("Body:", string(body))

	return body, err
}

func (rtc *RTC) requestXml(method string, url string, data string) (*models.Envelope, error) {
	resp, err := rtc.requestBody(method, url, data)
	if err != nil {
		return nil, err
	}

	env, err := models.NewFromXml(resp)
	return env, err
}

func (rtc *RTC) Login() error {
	url := fmt.Sprintf("https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/j_security_check?j_username=%s&j_password=%s", url.QueryEscape(rtc.User), url.QueryEscape(rtc.Password))
	resp, err := rtc.request("GET", url, "")

	if len(resp.Header["Location"]) > 0 && strings.Contains(resp.Header["Location"][0], "authfailed") {
		return errors.New("Failed to authenticate user " + rtc.User)
	}

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

	queryUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet?startIndex=0&maxResults=100&absoluteURIs=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&columnIdentifiers=workItemType&columnIdentifiers=summary&columnIdentifiers=creator&columnIdentifiers=owner&columnIdentifiers=creationDate&columnIdentifiers=duration&columnIdentifiers=category&columnIdentifiers=target&columnIdentifiers=projectArea&columnIdentifiers=internalTags&columnIdentifiers=internalState&itemId=_VMvycVRcEd61fuNW84kdiQ&skipOAuth=true&filterAttribute=&filterValue="

	resp, err := rtc.request("POST", queryUrl, "")
	if err != nil {
		return workItems, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	// fmt.Println("Body: ", string(body))

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
			State:        row.Labels[10],
			LocationUri:  row.LocationUri,
		}
		workItems = append(workItems, wi)
	}

	return workItems, err
}

type Filter struct {
	Field  string
	Oper   string
	Values []string
	Vars   []map[string]string
}

func (rtc *RTC) Query(filters []Filter, sortColumn string, sortAscending bool, maxResults int) ([]*WorkItem, error) {
	// curl "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet" -H "Cookie: com_ibm_team_process_web_ui_internal_admin_projects_ProcessTree_0SaveSelectedCookie="%"2F0; JazzFormAuth=Form; net-jazz-ajax-cookie-rememberUserId=; ibmSurvey=1422910922008; UnicaNIODID=r2adbtayyw2-ZDKlNvR; pSite=https"%"3A"%"2F"%"2Fwww.ibm.com"%"2Fdeveloperworks"%"2Ftopics"%"2Frest"%"2520api"%"2520"%"2520python"%"2F; mmcore.tst=0.911; mmid=-1314913985"%"7CAQAAAAo69LY+igsAAA"%"3D"%"3D; mmcore.pd=1780648624"%"7CAQAAAAoBQjr0tj6KC46Z2xgBAHt7sKJCDdJIEXd3dy5nb29nbGUuY29tLmJyDgAAAHt7sKJCDdJIAAAAAP////8AGQAAAP////8AEXd3dy5nb29nbGUuY29tLmJyBIoLAQAAAAAAAwAAAAAA////////////////AAAAAAABRQ"%"3D"%"3D; mmcore.srv=nycvwcgus02; CoreID6=79140352120814229109241&ci=50200000|DEVWRKS; CoreM_State=73~-1~-1~-1~-1~3~3~5~3~3~7~7~|~~|~~|~~|~||||||~|~~|~~|~~|~~|~~|~~|~~|~; CoreM_State_Content=6~|~~|~|; 50200000_clogin=v=1&l=1422910924&e=1422912724704; LtpaToken2=QT7AQ2NxDXkcEUJx0//EbS+Ta+y6IlVedjbU0yZvHSvf+W4Sxc7+s6iWWFrxE4hRkyvLTH7vrK3YQBJDUMVJSfDpv3v2AgOerm1oy/Vufc4fadGZdYiAdmIAwPIYnQpUIh30eY0EiSsXtPmxTbaOWEaniuAB5FeVy6SkYV/Ud6y2XR5UeXt0VuO+fcNNQM0ClosAE4Y3w9HgMGacuRfN3vNvh05yN87J3COyBb2m9RNcjpz0iY+YsaRxwJ7lZMPI3B5F+h9AREu5THQkczrcmVoUVwbB9bKdnIltP+nibQET5UXEzAh33tKaeKJ6Ivc3X2WkeIcxUHG1QCXTo1Jp8/uqUlaB+Fpl7TpnyLm7eFucKa3SqFiLA2Q3bsw+Cuuw95BWKsZmaHzc9bS+CJwevKREyAo2gcZMLsxouwl5daWm6LJkpvv1fLXfeOKiioNnuuA38262GLRCSVLsNYZatuuN00TRdzFQyjkYcH5uO5hHu0Od3mq+N+D8PfzWL4mrH9MrAi4CBf58mNA0NTri127jigDOcqRYdG8VZMOs+NLHxQGfJmZQZ3oQcrWP3phL3JrLKlb32OKu3tKDN2nxhR2ppiyKtK3uTVUOer5c0sbI6HUOtawD+VzyxHivgPLZg3sfwCZqD3+Z3uKd6KjalCFWPwqXKei3R3Zs1SjgXME=; JSESSIONID=0000w0b_QruvkcrpBiUBwiiWDew:-1" -H "X-jazz-downstream-auth-client-level: 4.0" -H "Origin: https://igartc01.swg.usma.ibm.com" -H "Accept-Encoding: gzip, deflate" -H "Accept-Language: en-US,en;q=0.8" -H "X-com-ibm-team-configuration-versions: LATEST" -H "User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36" -H "Content-Type: application/x-www-form-urlencoded; charset=UTF-8" -H "accept: text/json" -H "Referer: https://igartc01.swg.usma.ibm.com/jazz/web/projects/SD-OPS" -H "X-Requested-With: XMLHttpRequest" -H "Connection: keep-alive" --data "startIndex=0&maxResults=50&filterAttribute=&filterValue=&columnIdentifiers=workItemType&columnIdentifiers=id&columnIdentifiers=summary&columnIdentifiers=owner&columnIdentifiers=internalState&columnIdentifiers=internalPriority&columnIdentifiers=internalSeverity&columnIdentifiers=modified&sortColumns=modified&sortDirections=false&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&jsonExpression="%"7B"%"22operator"%"22"%"3A"%"22AND"%"22"%"2C"%"22attributeExpressions"%"22"%"3A"%"5B"%"7B"%"22attributeId"%"22"%"3A"%"22owner"%"22"%"2C"%"22operator"%"22"%"3A"%"22is"%"22"%"2C"%"22values"%"22"%"3A"%"5B"%"22_PrOIoMZ5Ed-Lr-wDR3V_pA"%"22"%"5D"%"2C"%"22variables"%"22"%"3A"%"5B"%"5D"%"7D"%"5D"%"2C"%"22termExpressions"%"22"%"3A"%"5B"%"5D"%"2C"%"22similarityExpressions"%"22"%"3A"%"5B"%"5D"%"7D" --compressed

	var workItems []*WorkItem

	// filter := Filter{Field: "owner", Oper: "is", Values: []string{"_PrOIoMZ5Ed-Lr-wDR3V_pA"}}
	// filters := []Filter{filter}

	mf := make(map[string]interface{})
	mf["operator"] = "AND"
	mf["termExpressions"] = []string{}
	mf["similarityExpressions"] = []string{}

	attrExps := []map[string]interface{}{}

	for _, f := range filters {
		attrExp := make(map[string]interface{})
		attrExp["attributeId"] = f.Field
		attrExp["operator"] = f.Oper
		attrExp["values"] = f.Values
		attrExp["variables"] = f.Vars

		attrExps = append(attrExps, attrExp)
	}

	mf["attributeExpressions"] = attrExps

	jsonStr, err := json.Marshal(mf)
	if err != nil {
		return workItems, err
	}

	queryUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet"
	// data := fmt.Sprintf("startIndex=0&maxResults=50&filterAttribute=&filterValue=&columnIdentifiers=workItemType&columnIdentifiers=id&columnIdentifiers=summary&columnIdentifiers=owner&columnIdentifiers=internalState&columnIdentifiers=internalPriority&columnIdentifiers=internalSeverity&columnIdentifiers=modified&sortColumns=modified&sortDirections=false&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&jsonExpression=%s", string(jsonStr))
	data := fmt.Sprintf("startIndex=0&maxResults=%d&filterAttribute=&filterValue=&columnIdentifiers=workItemType&Q&columnIdentifiers=summary&columnIdentifiers=creator&columnIdentifiers=owner&columnIdentifiers=creationDate&columnIdentifiers=duration&columnIdentifiers=category&columnIdentifiers=target&columnIdentifiers=projectArea&columnIdentifiers=internalTags&columnIdentifiers=internalState&sortColumns=%s&sortDirections=%t&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&jsonExpression=%s", maxResults, sortColumn, sortAscending, string(jsonStr))

	// fmt.Println(string(jsonStr))
	// fmt.Println(string(data))

	body, err := rtc.requestBody("POST", queryUrl, data)
	if err != nil {
		return workItems, err
	}

	// fmt.Println(string(body))

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
			State:        row.Labels[10],
			LocationUri:  row.LocationUri,
		}
		workItems = append(workItems, wi)
	}

	return workItems, nil
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
			State:       twi.StateName,
			PlannedFor:  "-",
		}
		workItems = append(workItems, wi)
	}

	return workItems, nil
}

func (rtc *RTC) GetInternalId(id string) (itemId string, stateId string, err error) {
	url := fmt.Sprintf("https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItemDTO2?includeHistory=false&id=%s&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", id)

	body, err := rtc.requestBody("GET", url, "")
	if err != nil {
		return "", "", err
	}

	// fmt.Println("Body", string(body))

	env, err := models.NewFromXml(body)
	if err != nil {
		return "", "", err
	}

	val := env.Body.Response.ReturnValue.Value

	return val.ItemId, val.StateId, nil
}

func (rtc *RTC) CreateNewId(wiType string) (string, error) {
	url := fmt.Sprintf("https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItemDTO2?includeHistory=false&newWorkItem=true&typeId=%s&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", wiType)

	env, err := rtc.requestXml("GET", url, "")
	if err != nil {
		return "", err
	}

	return env.Body.Response.ReturnValue.Value.ItemId, nil
}

func (rtc *RTC) Retrieve(id string) (*WorkItem, error) {
	retrieveUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/results?id=" + id + "&scopeToProject=false&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ"

	env, err := rtc.requestXml("GET", retrieveUrl, "")
	if err != nil {
		return nil, err
	}

	if len(env.Body.Response.ReturnValue.Value.WorkItems) < 1 {
		return nil, errors.New("No workitem with id " + id + " found")
	}

	rwi := env.Body.Response.ReturnValue.Value.WorkItems[0]

	wi := &WorkItem{
		Id:          rwi.Id,
		ItemId:      rwi.ItemId,
		Summary:     rwi.Summary,
		Type:        rwi.Type,
		OwnedBy:     rwi.OwnerName,
		CreatedBy:   rwi.CreatorName,
		LocationUri: rwi.LocationUri,
		Description: rwi.Description,
	}

	return wi, nil
}

func (rtc *RTC) Create(wi *WorkItem) (*WorkItem, error) {
	wiType := strings.ToLower(wi.Type)
	itemId, err := rtc.CreateNewId(wiType)

	if err != nil {
		return nil, err
	}

	var values map[string]string
	values = make(map[string]string)

	values["summary"] = url.QueryEscape(wi.Summary)
	values["workItemType"] = url.QueryEscape(wiType)
	values["work_product_where_found"] = "Work_Product_where_found.literal.l2"
	values["owner"] = rtc.OwnerId

	createUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItem2"
	data := fmt.Sprintf("itemId=%s&type=task&additionalSaveParameters=com.ibm.team.workitem.common.internal.updateBacklinks&sanitizeHTML=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", itemId)

	for k, _ := range values {
		data = data + "&attributeIdentifiers=" + k // url.QueryEscape(k)
		// fmt.Printf("%s => %s\n", k, url.QueryEscape(k))
	}

	for _, v := range values {
		data = data + "&attributeValues=" + v // url.QueryEscape(v)
		// fmt.Printf("%s => %s\n", v, url.QueryEscape(v))
	}

	// fmt.Println("\n\nTrying URL:", createUrl, data)
	// fmt.Println("\n")

	// return nil, nil
	env, err := rtc.requestXml("POST", createUrl, data)
	if err != nil {
		return nil, err
	}

	id := env.Body.Response.ReturnValue.Value.WorkItem.ParsedId()

	rwi, err := rtc.Retrieve(id)
	if err != nil {
		return nil, err
	}

	return rwi, nil
}

func (rtc *RTC) Request(method string, url string) ([]byte, error) {
	resp, err := rtc.requestBody(method, url, "")
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (rtc *RTC) GetReleases() ([]models.Release, error) {
	iterationsUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.process.internal.service.web.IProcessWebUIService/iterations?uuid=_nrtnkGAiEd6QQ7s7cowCAg"

	env, err := rtc.requestXml("GET", iterationsUrl, "")
	if err != nil {
		return nil, err
	}

	return env.Body.Response.ReturnValue.Releases, nil
}

func (rtc *RTC) GetIterations() ([]models.Iteration, error) {
	rels, err := rtc.GetReleases()
	if err != nil {
		return nil, err
	}

	var m []models.Iteration

	for _, rel := range rels {
		for _, iter := range rel.Iterations {
			m = append(m, iter)
		}
	}

	return m, nil
}

func (rtc *RTC) GetIterationsMap() (map[string]models.Iteration, error) {
	rels, err := rtc.GetReleases()
	if err != nil {
		return nil, err
	}

	var m map[string]models.Iteration
	m = make(map[string]models.Iteration)

	for _, rel := range rels {
		// fmt.Println("Release:", rel.Label)
		// fmt.Println("Iterations:", len(rel.Iterations))
		for _, iter := range rel.Iterations {
			// fmt.Println("   -", iter.ItemId, "-", iter.Label)
			m[iter.ItemId] = iter
		}
	}

	return m, nil
}

type Attributed interface {
	GetAttributes() []*models.Attribute
}

func getAttributes(a Attributed) map[string]string {
	var m map[string]string
	m = make(map[string]string)

	for _, attr := range a.GetAttributes() {
		m[attr.Key] = attr.Value.Label
		if m[attr.Key] == "" {
			m[attr.Key] = attr.Value.Content
		}
	}

	return m
}

func makeRef(link models.Link) Reference {
	refAttrs := getAttributes(link.Target)
	ref := Reference{
		ItemId:      link.Target.ItemId,
		StateId:     link.Target.StateId,
		LocationUri: link.Target.LocationUri,
		Summary:     refAttrs["summary"],
		Type:        refAttrs["workItemType"],
		Id:          refAttrs["id"],
	}

	return ref
}

func (rtc *RTC) GetWorkItem(id string) (*WorkItem, error) {
	wi, err := rtc.Retrieve(id)
	if err != nil {
		return nil, err
	}

	workItemUrl := fmt.Sprintf("https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItemDTO2?includeHistory=false&id=%s&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", id)
	env, err := rtc.requestXml("GET", workItemUrl, "")
	if err != nil {
		return nil, err
	}

	val := env.Body.Response.ReturnValue.Value
	attrs := getAttributes(val)

	wi.ItemId = val.ItemId
	wi.StateId = val.StateId
	wi.PlannedFor = attrs["target"]
	wi.State = attrs["internalState"]
	wi.Resolution = attrs["internalResolution"]
	wi.TimeSpent = attrs["timeSpent"]
	wi.Estimate = attrs["duration"]
	wi.CodeChanges = attrs["code-change"]

	// add parents
	if len(val.LinkTypes) > 0 {
		for _, lt := range val.LinkTypes {

			if lt.EndpointId == "parent" {
				for _, link := range lt.Links {
					ref := makeRef(link)
					wi.Parents = append(wi.Parents, ref)
				}
			}

			if lt.EndpointId == "children" {
				for _, link := range lt.Links {
					ref := makeRef(link)
					wi.Children = append(wi.Children, ref)
				}
			}
		}
	}

	// add approvals
	wi.Approvals = val.Approvals

	return wi, nil
}

func (rtc *RTC) ChangeStatus(id string, s string) (*models.Envelope, error) {
	itemId, stateId, err := rtc.GetInternalId(id)
	if err != nil {
		return nil, err
	}

	changeUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItem2"
	data := fmt.Sprintf("attributeIdentifiers=internalResolution&attributeValues=&action=bugzillaWorkflow.action.%s&itemId=%s&type=task&stateId=%s&additionalSaveParameters=com.ibm.team.workitem.common.internal.updateBacklinks&sanitizeHTML=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", s, itemId, stateId)
	env, err := rtc.requestXml("POST", changeUrl, data)
	if err != nil {
		return nil, err
	}

	return env, err
}

func (rtc *RTC) PerformAction(name string, id string, action string, expectedState string) error {
	env, err := rtc.ChangeStatus(id, action)
	if err != nil {
		return err
	}

	attrs := getAttributes(env.Body.Response.ReturnValue.Value.WorkItem)

	if state, ok := attrs["internalState"]; ok {
		if state == expectedState {
			return nil
		}
	}

	wi, err := rtc.GetWorkItem(id)
	if err != nil {
		return err
	}

	return errors.New("Failed to " + name + " work item " + id + ". Current state is " + wi.State + ".")
}

func (rtc *RTC) Close(id string) error {
	env, err := rtc.ChangeStatus(id, "close")
	if err != nil {
		return err
	}

	attrs := getAttributes(env.Body.Response.ReturnValue.Value.WorkItem)

	if state, ok := attrs["internalState"]; ok {
		if state == "Closed" {
			return nil
		}
	}

	wi, err := rtc.GetWorkItem(id)
	if err != nil {
		return err
	}

	return errors.New("Failed to close work item " + id + ". Current state is " + wi.State + ".")
}

func (rtc *RTC) SetAttributes(id string, attrs map[string]string) (*models.Envelope, error) {
	itemId, stateId, err := rtc.GetInternalId(id)
	if err != nil {
		return nil, err
	}

	changeUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItem2"
	data := fmt.Sprintf("type=task&itemId=%s&stateId=%s&additionalSaveParameters=com.ibm.team.workitem.common.internal.updateBacklinks&sanitizeHTML=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", itemId, stateId)

	for k, v := range attrs {
		data += fmt.Sprintf("&attributeIdentifiers=%s&attributeValues=%s", k, v)
	}

	env, err := rtc.requestXml("POST", changeUrl, data)
	if err != nil {
		return nil, err
	}

	return env, err
}

func (rtc *RTC) SaveAttribute(id string, attr string, value string) (*models.Envelope, error) {
	itemId, stateId, err := rtc.GetInternalId(id)
	if err != nil {
		return nil, err
	}

	changeUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItem2"
	data := fmt.Sprintf("attributeIdentifiers=%s&attributeValues=%s&itemId=%s&type=task&stateId=%s&additionalSaveParameters=com.ibm.team.workitem.common.internal.updateBacklinks&sanitizeHTML=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", attr, value, itemId, stateId)

	env, err := rtc.requestXml("POST", changeUrl, data)
	if err != nil {
		return nil, err
	}

	return env, err
}

func (rtc *RTC) AddParent(id string, parentId string) error {
	wi, err := rtc.GetWorkItem(id)
	if err != nil {
		return err
	}

	pwi, err := rtc.GetWorkItem(parentId)
	if err != nil {
		return err
	}

	link := url.QueryEscape(fmt.Sprintf(`{"cmd":"addLink","type":"com.ibm.team.workitem.linktype.parentworkitem","end":"target","name":"Parent","itemId":"%s","comment":"%s: %s"}`, pwi.ItemId, pwi.Id, pwi.Summary))

	linkUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItem2"
	data := fmt.Sprintf("itemId=%s&type=task&stateId=%s&updateLinks=%s&additionalSaveParameters=com.ibm.team.workitem.common.internal.updateBacklinks&sanitizeHTML=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ", wi.ItemId, wi.StateId, link)

	_, err = rtc.requestBody("POST", linkUrl, data)
	if err != nil {
		return err
	}

	return nil
}

func (rtc *RTC) Update(wi WorkItem) error {
	m := make(map[string]string)

	if wi.TimeSpent != "" {
		m["timeSpent"] = wi.TimeSpent
	}

	if wi.Estimate != "" {
		m["duration"] = wi.Estimate
	}

	if wi.IterationId != "" {
		iters, err := rtc.GetIterations()
		if err != nil {
			return err
		}

		iterIdNum, err := strconv.Atoi(wi.IterationId)

		if iterIdNum > len(iters) {
			return errors.New("Iteration with id " + wi.IterationId + " not found. Use iterations command.")
		}

		iter := iters[iterIdNum]
		m["target"] = iter.ItemId
	}

	if wi.Id == "" {
		return errors.New("Missing work item id")
	}

	_, err := rtc.SetAttributes(wi.Id, m)
	return err
}

func (rtc *RTC) MoveToIteration(id string, iterId string) (*WorkItem, models.Iteration, error) {
	var iter models.Iteration

	wi, err := rtc.GetWorkItem(id)
	if err != nil {
		return nil, iter, err
	}

	if wi == nil {
		return nil, iter, errors.New("Work item with id " + id + " not found. Use list command.")
	}

	iters, err := rtc.GetIterations()
	if err != nil {
		return nil, iter, err
	}

	iterIdNum, err := strconv.Atoi(iterId)

	if iterIdNum > len(iters) {
		return nil, iter, errors.New("Iteration with id " + iterId + " not found. Use iterations command.")
	}

	iter = iters[iterIdNum]
	// fmt.Println("Iter:", iter.ItemId, "-", iter.Label)

	_, err = rtc.SaveAttribute(id, "target", iter.ItemId)
	if err != nil {
		return nil, iter, err
	}

	return wi, iter, nil
}

func (rtc *RTC) CreateSubTask(id string, subType string) (*WorkItem, error) {
	pwi, err := rtc.Retrieve(id)
	if err != nil {
		return nil, err
	}

	summary := fmt.Sprintf("%s: %s", subType, pwi.Summary)

	wi := &WorkItem{
		Summary: summary,
		Type:    "task",
	}

	wi, err = rtc.Create(wi)
	if err != nil {
		return nil, err
	}

	err = rtc.AddParent(wi.Id, pwi.Id)
	if err != nil {
		return nil, err
	}

	return wi, nil
}

func (rtc *RTC) GetAllValues() (map[string]map[string]string, error) {
	allValuesUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/allValues?projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&typeId=task&includeArchived=false&ids=workItemType&ids=internalSeverity&ids=foundIn&ids=creator&ids=category&ids=internalTags&ids=internalPriority&ids=owner&ids=target&ids=task&ids=key-component&ids=environment&itemId=_IDpPV6fhEeSicYpAbHXWsw"

	env, err := rtc.requestXml("GET", allValuesUrl, "")
	if err != nil {
		return nil, err
	}

	return env.Body.Response.ReturnValue.GetItems(), nil
}

func (rtc *RTC) GetOwners() ([]Owner, error) {
	values, err := rtc.GetAllValues()
	owners := []Owner{}
	for k, v := range values["owner"] {
		owners = append(owners, Owner{Id: k, Name: v})
	}

	sort.Sort(byName(owners))
	return owners, err
}

func (rtc *RTC) OpenWorkItem(id string) error {
	wi, err := rtc.GetWorkItem(id)
	if err != nil {
		return err
	}

	return open.Start(wi.LocationUri)
}

func (rtc *RTC) AddApproval(id string, desc string, ownerId string) error {
	itemId, stateId, err := rtc.GetInternalId(id)
	if err != nil {
		return err
	}

	createUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItem2"
	approval := fmt.Sprintf(`{"cmd":"createApproval","param":{"type":"com.ibm.team.workitem.approvalType.approval","name":"%s","approvers":[{"user":"%s","state":"com.ibm.team.workitem.approvalState.pending"}]}}`, desc, ownerId)
	data := fmt.Sprintf("attributeIdentifiers=internalResolution&attributeValues=&itemId=%s&type=task&stateId=%s&updateApprovals=%s", itemId, stateId, approval)

	_, err = rtc.requestXml("POST", createUrl, data)
	if err != nil {
		return err
	}

	return nil
}

// values["category"] = "_aXl2IGW0Ed6uZsIllQzRvg"
// values["owner"] = "_PrOIoMZ5Ed-Lr-wDR3V_pA"
// values["target"] = "_H5fMQaHXEeS4fen3HD7Mow"
// values["documentation-changes"] = ""
// values["process"] = "component.literal.l13"
// values["comments"] = ""
// values["com.ibm.team.apt.estimate.minimal"] = ""
// values["com.ibm.team.apt.estimate.maximal"] = ""
// values["items-for-deployment"] = ""
// values["package"] = "build-package-type.literal.l1"
// values["was"] = ""
// values["mq"] = ""
// values["code-review-time"] = "0"
// values["root"] = ""
// values["task"] = "task-type.literal.l1"
// values["key-component"] = "component-real.literal.l21"
// values["properties-change"] = ""
// values["code-change"] = ""
// values["database-change"] = ""
// values["test-state"] = "test-state.literal.l1"
// values["testPhase"] = "test-phase.literal.l7"
// values["defectType"] = "defect-type.literal.l11"
// values["environment"] = "environment.literal.l9"
// values["defectReproduced"] = "defect-reproduced.literal.l1"
// values["phaseDetected"] = "phase-detected.literal.l7"
// values["projectArea"] = "_U7zMYFRcEd61fuNW84kdiQ"
// values["internalTags"] = ""
// values["internalPriority"] = "priority.literal.l01"
// values["timeSpent"] = ""
// values["duration"] = ""
// values["correctedEstimate"] = ""
// values["internalSequenceValue"] = ""
// values["description"] = ""
// values["internalResolution"] = "null"
// values["internalSeverity"] = "severity.literal.l3"
// values["contextId"] = "_U7zMYFRcEd61fuNW84kdiQ"
// values["archived"] = "false"
// values["process_root_cause"] = "Process_Root_Cause.literal.l27"
// values["work_product_where_found"] = "Work_Product_where_found.literal.l2"
// values["phase_injected"] = "Phase_Injected.literal.l7"
// values["phase_detected"] = "Phase_Detected.literal.l13"
// values["defect_type"] = "Defect_Type.literal.l10"
// values["process_area_where_found"] = "Process_Area_where_found.literal.l45"
// values["activity_found"] = "Activity_Found.literal.l45"
// values["teamArea"] = "_1Hjb4FYBEd69G8biF2ewvg"
// values["stateGroup"] = "0"

// curl "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IWorkItemRestService/workItemDTO2?includeHistory=false&id=1279685" -H "Cookie: com_ibm_team_process_web_ui_internal_admin_projects_ProcessTree_0SaveSelectedCookie="%"2F0; JazzFormAuth=Form; net-jazz-ajax-cookie-rememberUserId=; ibmSurvey=1422910922008; UnicaNIODID=r2adbtayyw2-ZDKlNvR; pSite=https"%"3A"%"2F"%"2Fwww.ibm.com"%"2Fdeveloperworks"%"2Ftopics"%"2Frest"%"2520api"%"2520"%"2520python"%"2F; mmcore.tst=0.911; mmid=-1314913985"%"7CAQAAAAo69LY+igsAAA"%"3D"%"3D; mmcore.pd=1780648624"%"7CAQAAAAoBQjr0tj6KC46Z2xgBAHt7sKJCDdJIEXd3dy5nb29nbGUuY29tLmJyDgAAAHt7sKJCDdJIAAAAAP////8AGQAAAP////8AEXd3dy5nb29nbGUuY29tLmJyBIoLAQAAAAAAAwAAAAAA////////////////AAAAAAABRQ"%"3D"%"3D; mmcore.srv=nycvwcgus02; CoreID6=79140352120814229109241&ci=50200000|DEVWRKS; CoreM_State=73~-1~-1~-1~-1~3~3~5~3~3~7~7~|~~|~~|~~|~||||||~|~~|~~|~~|~~|~~|~~|~~|~; CoreM_State_Content=6~|~~|~|; 50200000_clogin=v=1&l=1422910924&e=1422912724704; LtpaToken2=RJK9zbo12q7pw1YkMXfWOU51+KfbWoyeT/ch7qEqU/l3nrGfBPnbNhqcsMREiueSJoYh7x6q6z0YHat8nquCcCjIz4R/bI/QCq14jALWJLny7odPnhl6FzOcTiRQKOAq6UzYblOVjJvvN6/UaQCVTIe66vVS0/hH6B9z8hpwhMZnTNARAZkjZdba2B8+5erfi0OBpL/HSsyXxGLRoSuemvqAmMvV0WObDNvxpzba9PrmpvdNPD9sNFw1UOy16Vq6I77ZEfbY7AHB/nbprFIdjcM90T7rcSQChDgmxqumlwASjX1iUNvqxatLzdeKa+Lb7uHsrM5apxo3Q+PvWkDD7SzkZPYqq+CXhKaVnOmPvq93DZcvUQMJ9nEbGK/uFu5tfi44bGr3NF4b9pUmJD9H2K8O1KmGXQzzu8/EtiKnL0ANdPcXALU983uT3NdWKGJdWS1LN/CsTg54MrG/Wx0hGPcz8SJPXNUuF3OTPFRO/WRdahYdUhOsYYvS0nOAgZzl9QStwpM9gg5maCQxvoEtMCihXuhb9yggLUf48NS9ei+9E5oSmFRMhAsVALoMTvHvVhCLINN5sNRw1ouvdNdOArcIXf/gNZQHDdcF7lheYWRWcFF4abvLyBjuFdxh6Moso/KzMkHy4mNmSkcTVhem4hap/kyQ7PN3T3likGo6PXk=; JSESSIONID=0000McnEpThigYe4KZtmvEkcYTo:-1" -H "X-jazz-downstream-auth-client-level: 4.0" -H "Accept-Encoding: gzip, deflate, sdch" -H "Accept-Language: en-US,en;q=0.8" -H "X-com-ibm-team-configuration-versions: LATEST" -H "User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36" -H "Content-Type: application/x-www-form-urlencoded; charset=utf-8" -H "accept: text/json" -H "Referer: https://igartc01.swg.usma.ibm.com/jazz/web/projects/SD-OPS" -H "X-Requested-With: XMLHttpRequest" -H "Connection: keep-alive" --compressed
