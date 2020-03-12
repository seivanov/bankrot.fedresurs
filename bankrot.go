package main

import (
    "fmt"
    "net/http"
    "bytes"
    "regexp"
    "time"
    "strings"
    "strconv"
    "os"
    "bufio"
    "github.com/PuerkitoBio/goquery"
    "encoding/csv"
)

var writer *csv.Writer
var list []string

func readList() []string {

    file, err := os.Open("list.txt")

    if err != nil {
    }

    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    var txtlines []string

    for scanner.Scan() {
        txtlines = append(txtlines, scanner.Text())
    }

    file.Close()

    return txtlines

}

func eq(str string) bool {

    str = strings.Replace(str, "&#34;", "", -1)
    str = strings.Replace(str, "«", "", -1)
    str = strings.Replace(str, "»", "", -1)

    str = strings.TrimSpace(str)

    for _, element := range list {

        element = strings.Trim(element, "&#34;")
        element = strings.TrimSpace(element)
        element = strings.Replace(element, "«", "", -1)
        element = strings.Replace(element, "»", "", -1)

        if strings.Compare(str, element) == 0 {
            fmt.Println("-- >" + str + "<")
            return false
        }
    }

    fmt.Println("++ >" + str + "<")
    return true
}

func getContent(url string) {

    r5, _ := regexp.Compile(`MessageWindow.aspx`)
    p5 := r5.FindStringSubmatch(url)

    if len(p5) == 0 {
        return
    }

    fmt.Println("> " + url)

    var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    if err != nil {
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
    }

    var number string
    var date string
    var fio string
    var inn string
    var live string
    var place string

    doc.Find("table tr").Each(func(i int, s *goquery.Selection) {

        col1, _ := s.Find("td").Eq(0).Html()
        col1 = strings.TrimSpace(col1)
        col2, _ := s.Find("td").Eq(1).Html()
        col2 = strings.TrimSpace(col2)

        if col1 == "Дата публикации" {
            date = col2
        }

        if col1 == "№ сообщения" {
            number = col2
        }

        if col1 == "ФИО должника" || col1 == "Наименование должника" {
            fio = col2
        }

        if col1 == "ИНН" {
            inn = col2
        }

        if col1 == "Место жительства" {
            live = col2
        }

        if col1 == "Место проведения:" {
            place = col2
        }

    })

    //fmt.Println(number + " " + fio + " " + inn + " " + live + " " + place)

    if eq(place) {
        writer.Write([]string{number, date, fio, inn, live, place})
        writer.Flush()
    }

}

func main() {

    list = readList()

    t := time.Now().Add(-24*time.Hour)
    date_start := t.Format("02.01.2006")
    t = time.Now()
    date_end := t.Format("02.01.2006")

    fmt.Println(date_start + " " + date_end)

    file, _ := os.Create("result.csv")
    defer file.Close()

    bomUtf8 := []byte{0xEF, 0xBB, 0xBF}
    file.Write(bomUtf8)

    writer = csv.NewWriter(file)
    writer.Comma = ';'
    defer writer.Flush()

    var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
    var Page = "0"

    for {

        fmt.Println("%%%%%%%%%%%>"+Page+"<%%%%%%%%%%")

        if Page == "" {
            break
        }

        req, err := http.NewRequest("POST", "https://bankrot.fedresurs.ru/Messages.aspx", bytes.NewBuffer(jsonStr))
        req.Header.Set("Cookie", "Messages=MessageNumber=&MessageType=Auction&MessageTypeText=%d0%9e%d0%b1%d1%8a%d1%8f%d0%b2%d0%bb%d0%b5%d0%bd%d0%b8%d0%b5+%d0%be+%d0%bf%d1%80%d0%be%d0%b2%d0%b5%d0%b4%d0%b5%d0%bd%d0%b8%d0%b8+%d1%82%d0%be%d1%80%d0%b3%d0%be%d0%b2&DateEndValue="+date_end+"+0%3a00%3a00&DateBeginValue="+date_start+"+0%3a00%3a00&PageNumber="+Page+"&DebtorText=&DebtorId=&DebtorType=&PublisherType=&PublisherId=&PublisherText=&IdRegion=&IdCourtDecisionType=&WithAu=False&WithViolation=False; _ym_visorc_45311283=w; AmListSearch=SroId=&SroName=&FirstName=&LastName=&MiddleName=&RegNumber=&PageNumber="+Page+"&WithPublicatedMessages=False")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            //panic(err)
            continue
        }
        defer resp.Body.Close()

        //fmt.Println("response Status:", resp.Status)
        //fmt.Println("response Headers:", resp.Header)
        //body, _ := ioutil.ReadAll(resp.Body)
        //fmt.Println("response Body:", string(body))

        doc, err := goquery.NewDocumentFromReader(resp.Body)
        if err != nil {
        }

        doc.Find("table.bank tr").Each(func(i int, s *goquery.Selection) {
            var href, _ = s.Find("td").Eq(1).Find("a").Attr("href")
            //var title, _ = s.Find("td").Eq(2).Find("a").Attr("title")

            //fmt.Println(href + " " + title)
            //fmt.Println(href)

            getContent("https://bankrot.fedresurs.ru" + href)

        })

        r5, _ := regexp.Compile(`Page\$([0-9]+)`)

        newpage := ""

        var find = false
        doc.Find(".pager table tbody tr td").EachWithBreak(func(i int, s *goquery.Selection) bool {

            if s.Find("a").Length() == 0 {
                find = true
            } else if find == true {

                //var Page = s.Find("a").Text()

                var href, _ = s.Find("a").Attr("href")
                p5 := r5.FindStringSubmatch(href)
                p_int, _ := strconv.Atoi(p5[1])
                p_int -= 1
                newpage = strconv.Itoa(p_int)

                fmt.Println("Page:" + newpage)

                return false
            }

            return true

        })

        if strings.Compare(Page, newpage) == 0 {
            break
        } else {
            Page = newpage
        }

        time.Sleep(time.Second)

    }

}
