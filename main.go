package main

import (
	"fmt"
	"io/ioutil"

	"github.com/fcoury/rtc-go/browser"
	"github.com/fcoury/rtc-go/models"
)

// func Authenticate() (http.Response, error) {
// 	loginRequestUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/j_security_check?j_username=fcoury%40br.ibm.com&j_password=tempra14"
// 	// loginRequestUrl := "http://requestb.in/qkc22pqk"

// 	r, err := http.NewRequest("GET", loginRequestUrl, nil) // bytes.NewBufferString("j_username=fcoury%40br.ibm.com&j_password=tempra14"))

// 	if err != nil {
// 		return err
// 	}

// 	r.Header.Add("Host", "igartc01.swg.usma.ibm.com")
// 	r.Header.Add("Connection", "keep-alive")
// 	// r.Header.Add("Content-Length", "50")
// 	r.Header.Add("X-jazz-downstream-auth-client-level", "4.0")
// 	r.Header.Add("Origin", "https://igartc01.swg.usma.ibm.com")
// 	r.Header.Add("X-Requested-With", "XMLHttpRequest")
// 	r.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36")
// 	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
// 	r.Header.Add("Accept", "*/*")
// 	r.Header.Add("Referer", "https://igartc01.swg.usma.ibm.com/jazz/web/projects/SD-OPS")
// 	r.Header.Add("Accept-Encoding", "gzip, deflate")
// 	r.Header.Add("Accept-Language", "en-US,en;q=0.8")
// 	r.Header.Add("Cookie", "JazzFormAuth=Form; net-jazz-ajax-cookie-rememberUserId=; WASReqURL=https:///jazz/authenticated/identity?redirectPath=%252Fjazz%252Fservice%252Fcom.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService%252FinitializationData")

// 	resp, err := http.DefaultTransport.RoundTrip(r)

// 	if err != nil {
// 		return err
// 	}

// 	return resp
// }

// func CreateSession(url string) (http.Response, error) {
// 	loginRequestUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/j_security_check?j_username=fcoury%40br.ibm.com&j_password=tempra14"

// 	r, err := http.NewRequest("GET", loginRequestUrl, nil)

// 	if err != nil {
// 		return nil, err
// 	}

// 	for i := range resp.Cookies() {
// 		cookie := resp.Cookies()[i]
// 		r.AddCookie(cookie)
// 	}

// 	r.Header.Add("Cookie", resp.Header["Set-Cookie"][0])

// 	resp, err = http.DefaultTransport.RoundTrip(r)

// }

func main() {

	// xml, err := ioutil.ReadFile("sample.xml")

	b := browser.NewBrowser()
	_, err := b.Request("GET", "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/j_security_check?j_username=fcoury%40br.ibm.com&j_password=tempra14", "")

	if err != nil {
		panic(err)
	}

	_, err = b.Request("POST", "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/", "")

	if err != nil {
		panic(err)
	}

	data := "startIndex=0&maxResults=100&absoluteURIs=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&columnIdentifiers=workItemType&columnIdentifiers=summary&columnIdentifiers=creator&columnIdentifiers=owner&columnIdentifiers=creationDate&columnIdentifiers=duration&columnIdentifiers=category&columnIdentifiers=target&columnIdentifiers=projectArea&columnIdentifiers=internalTags&itemId=_VMvycVRcEd61fuNW84kdiQ&skipOAuth=true&filterAttribute=&filterValue="
	queryUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet"

	resp, err := b.Request("POST", queryUrl, data)

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	env, err := models.NewFromXml(body)

	if err != nil {
		panic(err)
	}

	// fmt.Printf("Data: %+v\n", env)

	for _, row := range env.Body.Response.ReturnValue.Value.Rows {
		fmt.Printf("Item: %d - %s\n", row.Id, row.Labels[1])
	}

	// fmt.Println("Requesting...")

	// b := browser.NewBrowser()
	// _, err := b.Request("GET", "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/j_security_check?j_username=fcoury%40br.ibm.com&j_password=tempra14", "")

	// if err != nil {
	// 	panic(err)
	// }

	// _, err = b.Request("POST", "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/", "")

	// if err != nil {
	// 	panic(err)
	// }

	// data := "startIndex=0&maxResults=10&absoluteURIs=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&columnIdentifiers=workItemType&columnIdentifiers=summary&columnIdentifiers=creator&columnIdentifiers=owner&columnIdentifiers=creationDate&columnIdentifiers=duration&columnIdentifiers=category&columnIdentifiers=target&columnIdentifiers=projectArea&columnIdentifiers=internalTags&itemId=_VMvycVRcEd61fuNW84kdiQ&skipOAuth=true&filterAttribute=&filterValue="
	// queryUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet"

	// resp, err := b.Request("POST", queryUrl, data)

	// if err != nil {
	// 	panic(err)
	// }

	// body, err := ioutil.ReadAll(resp.Body)

	// res, err := models.NewFromXml(string(body))

	// // fmt.Printf("Data: %s\n", string(body))
	// fmt.Printf("Data: %+v\n", res.Body)

	// cookies = resp.Cookies()

	// resp, err = CreateSession(resp.Header["Location"][0])

	// // loginUrl := "https://igartc01.swg.usma.ibm.com/jazz/web/projects/SD-OPS"

	// // resp, err := http.Get(loginUrl)
	// // if err != nil {
	// // 	panic(err)
	// // }

	// // body, err := ioutil.ReadAll(resp.Body)

	// // if err != nil {
	// // 	panic(err)
	// // }

	// // fmt.Printf("Headers: %s\nBody: %s\n", resp.Header, body)

	// // curl -i --insecure "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/j_security_check" -H "Cookie: JazzFormAuth=Form; net-jazz-ajax-cookie-rememberUserId=; WASReqURL=https:///jazz/authenticated/identity?redirectPath="%"252Fjazz"%"252Fservice"%"252Fcom.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService"%"252FinitializationData" -H "X-jazz-downstream-auth-client-level: 4.0" -H "Origin: https://igartc01.swg.usma.ibm.com" -H "Accept-Encoding: gzip, deflate" -H "Accept-Language: en-US,en;q=0.8" -H "User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36" -H "Content-Type: application/x-www-form-urlencoded; charset=UTF-8" -H "Accept: */*" -H "Referer: https://igartc01.swg.usma.ibm.com/jazz/web/projects/SD-OPS" -H "X-Requested-With: XMLHttpRequest" -H "Connection: keep-alive" --data "j_username=fcoury"%"40br.ibm.com&j_password=tempra14" --compressed

	// // r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	// // tr := &http.Transport{
	// // 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// // }
	// // client := &http.Client{Transport: tr}
	// // client := &http.Client{}

	// // resp, err := client.Do(r)
	// resp, err := http.DefaultTransport.RoundTrip(r)

	// if err != nil {
	// 	panic(err)
	// }

	// // _, err := ioutil.ReadAll(resp.Body)

	// // if err != nil {
	// // 	panic(err)
	// // }

	// // fmt.Printf("Status: %s\nHeaders: %s\nBody: %s\n", resp.Status, resp.Header, body)

	// for k, v := range resp.Header {
	// 	fmt.Printf("%s = %s\n", k, v)
	// }

	// // curl "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.repository.service.internal.webuiInitializer.IWebUIInitializerRestService/" -H "Cookie: net-jazz-ajax-cookie-rememberUserId=; LtpaToken2=b8m+7bEPnviACPM0lq1j0EMghru5rJn4Qav+FxbHWB0aVTg2L1MV9zelhJP4/OVNHFZA/aiPi+5nAZitZPnUi8K63pf21HeB2AYRU4P5XIEf/Tps7woCDhZuQDEgXPORKdgOM+JA2LJRPSoTANkk4Ij7A3CWJnno7ZVJTjcNXY2fmMWAE0VhbIqOabbo9Oov15G1H3e4IAVp6wmb15UDsOhYExkthSdlGVPMvUCycCpOgzCCf2Axt4itAAp4PszJx4FYDpwVVzY9m3f29HFyONOuPWkju1aMohF801CYCBWVHS2UlhCkXmFcZv9xJTfafRPP1QPOqFXRehUnRGNwuldx+JRlioOIC2m2052t6jyvxBrEeGxTx3A78Vl63VPIfiPz7Ty3G/b7gL/a+bP9tVGIIDcvW7tfauzCKKH9+HIY43JDU3hrGXflHgnvaCfPbO5o5OKgaj3kX4XKEvAEd1O6dUL4IPiJVpWue0Ifr2CqUrnopCI2XsOWf0IcS2GbpF77Sp2bFViIm9mxckgnEH7hpp5Lvy7eBbC3mX0Io236fuka8LJQnrBySH3jHykLvdLAHP8yAX8fPPLEYNp18R1sLghhA/BEOHM58Rm+UztCe9zWRUmJL86mWwy1NoY1RWEMcbzPvvKOin551AOjQyW6mCZV50KR3Y7wrZRsWYQ=" -H "X-jazz-downstream-auth-client-level: 4.0" -H "Accept-Encoding: gzip, deflate, sdch" -H "Accept-Language: en-US,en;q=0.8" -H "User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36" -H "Accept: */*" -H "Referer: https://igartc01.swg.usma.ibm.com/jazz/web/projects/SD-OPS" -H "X-Requested-With: XMLHttpRequest" -H "Connection: keep-alive" --compressed

	// fmt.Printf("\n\n----\n%s\n----\n\n", resp.Cookies())

	// for i := range resp.Cookies() {
	// 	fmt.Printf("Cookie: %s\n", resp.Cookies()[i])
	// }

	// loginUrl := resp.Header["Location"][0]
	// r, _ = http.NewRequest("GET", loginUrl, nil)
	// for i := range resp.Cookies() {
	// 	cookie := resp.Cookies()[i]
	// 	r.AddCookie(cookie)
	// }

	// r.Header.Add("Cookie", resp.Header["Set-Cookie"][0])

	// resp, err = http.DefaultTransport.RoundTrip(r)
	// body, err := ioutil.ReadAll(resp.Body)

	// fmt.Printf("\n\nStatus: %s\nHeaders: %s\nBody: %s\n", resp.Status, resp.Header, body)

	// var jSessionCookie *http.Cookie
	// for i := range resp.Cookies() {
	// 	cookie := resp.Cookies()[i]
	// 	fmt.Printf("Cookie: %s\n", resp.Cookies()[i].Name)
	// 	if cookie.Name == "JSESSIONID" {
	// 		fmt.Println("FOUND")
	// 		jSessionCookie = cookie
	// 	}
	// }

	// data := "startIndex=0&maxResults=10&absoluteURIs=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&columnIdentifiers=workItemType&columnIdentifiers=summary&columnIdentifiers=creator&columnIdentifiers=owner&columnIdentifiers=creationDate&columnIdentifiers=duration&columnIdentifiers=category&columnIdentifiers=target&columnIdentifiers=projectArea&columnIdentifiers=internalTags&itemId=_VMvycVRcEd61fuNW84kdiQ&skipOAuth=true&filterAttribute=&filterValue="
	// queryUrl := "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet"
	// // queryUrl := "http://requestb.in/qkc22pqk"

	// fmt.Printf("jSessionCookie: %s\n", jSessionCookie)
	// r, _ = http.NewRequest("POST", queryUrl, bytes.NewBufferString(data))
	// for i := range resp.Cookies() {
	// 	cookie := resp.Cookies()[i]
	// 	r.AddCookie(cookie)
	// }

	// resp, err = http.DefaultTransport.RoundTrip(r)
	// body, err = ioutil.ReadAll(resp.Body)

	// fmt.Printf("\n\nStatus: %s\nHeaders: %s\nBody: %s\n", resp.Status, resp.Header, body)

	// // curl "https://igartc01.swg.usma.ibm.com/jazz/service/com.ibm.team.workitem.common.internal.rest.IQueryRestService/getResultSet" -H "Cookie: JazzFormAuth=Form; net-jazz-ajax-cookie-rememberUserId=; LtpaToken2=b8m+7bEPnviACPM0lq1j0EMghru5rJn4Qav+FxbHWB0aVTg2L1MV9zelhJP4/OVNHFZA/aiPi+5nAZitZPnUi8K63pf21HeB2AYRU4P5XIEf/Tps7woCDhZuQDEgXPORKdgOM+JA2LJRPSoTANkk4Ij7A3CWJnno7ZVJTjcNXY2fmMWAE0VhbIqOabbo9Oov15G1H3e4IAVp6wmb15UDsOhYExkthSdlGVPMvUCycCpOgzCCf2Axt4itAAp4PszJx4FYDpwVVzY9m3f29HFyONOuPWkju1aMohF801CYCBWVHS2UlhCkXmFcZv9xJTfafRPP1QPOqFXRehUnRGNwuldx+JRlioOIC2m2052t6jyvxBrEeGxTx3A78Vl63VPIfiPz7Ty3G/b7gL/a+bP9tVGIIDcvW7tfauzCKKH9+HIY43JDU3hrGXflHgnvaCfPbO5o5OKgaj3kX4XKEvAEd1O6dUL4IPiJVpWue0Ifr2CqUrnopCI2XsOWf0IcS2GbpF77Sp2bFViIm9mxckgnEH7hpp5Lvy7eBbC3mX0Io236fuka8LJQnrBySH3jHykLvdLAHP8yAX8fPPLEYNp18R1sLghhA/BEOHM58Rm+UztCe9zWRUmJL86mWwy1NoY1RWEMcbzPvvKOin551AOjQyW6mCZV50KR3Y7wrZRsWYQ=; JSESSIONID=0000DUdgkZvREbMD2WObEjs7AQU:-1" -H "X-jazz-downstream-auth-client-level: 4.0" -H "Origin: https://igartc01.swg.usma.ibm.com" -H "Accept-Encoding: gzip, deflate" -H "Accept-Language: en-US,en;q=0.8" -H "X-com-ibm-team-configuration-versions: LATEST" -H "User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36" -H "Content-Type: application/x-www-form-urlencoded; charset=UTF-8" -H "accept: text/json" -H "Referer: https://igartc01.swg.usma.ibm.com/jazz/web/projects/SD-OPS" -H "X-Requested-With: XMLHttpRequest" -H "Connection: keep-alive" --data "startIndex=0&maxResults=10&absoluteURIs=true&projectAreaItemId=_U7zMYFRcEd61fuNW84kdiQ&columnIdentifiers=workItemType&columnIdentifiers=summary&columnIdentifiers=creator&columnIdentifiers=owner&columnIdentifiers=creationDate&columnIdentifiers=duration&columnIdentifiers=category&columnIdentifiers=target&columnIdentifiers=projectArea&columnIdentifiers=internalTags&itemId=_VMvycVRcEd61fuNW84kdiQ&skipOAuth=true&filterAttribute=&filterValue=" --compressed
}
