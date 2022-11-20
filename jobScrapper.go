package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	ccsv "github.com/tsak/concurrent-csv-writer"
)

type extractedJob struct {
	title string
	corp string
	condition string
	location string
	link string
	tag string
}

var keyword string = os.Args[1]
var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?searchword=" + keyword +"&go=&flag=n&searchMode=1&searchType=search&search_done=y&search_optional_item=n&recruitPage=1&recruitSort=relation&recruitPageCount=100&inner_com_type=&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&show_applied=&quick_apply=&except_read=&ai_head_hunting="

func main() {
	
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages()
	// fmt.Println("totalPages : ", totalPages)
	
	for i := 1; i <= totalPages; i++ {
		go getPage(i, c)
	}
	
	start := time.Now()

	for i := 1; i<=totalPages; i++ {
		extractedJobs := <- c
		jobs = append(jobs, extractedJobs...)
	}
	
	// writeJobs(jobs)
	cwriteJobs2(jobs)
	elapsed := time.Since(start)
	fmt.Printf("The total time %s\n", elapsed)
	fmt.Println("Done, extracted", len(jobs))
}

// func writeJobs(jobs []extractedJob) {
// 	file, err := os.Create(keyword + ".csv")
// 	checkErr(err)

// 	w := csv.NewWriter(file)
// 	defer w.Flush()

// 	headers := []string{"Title", "Corp", "Condition", "Location", "Tag", "Link"}

// 	wErr := w.Write(headers)
// 	checkErr(wErr)

// 	// write jobs
// 	for _, job := range jobs {
// 		jobSlice := []string{job.title, job.corp, job.condition, job.location, job.tag, job.link}
// 		jwErr := w.Write(jobSlice)
// 		checkErr(jwErr)
// 	}
// }

func cwriteJobs2(jobs []extractedJob) {
	csv, err := ccsv.NewCsvWriter(keyword + ".csv")
	checkErr(err)

	defer csv.Close()

	headers := []string{"Title", "Corp", "Condition", "Location", "Tag", "Link"}

	csv.Write(headers)

	done := make(chan bool)

	for _, job := range jobs {
		go func(job extractedJob) {
			csv.Write([]string{job.title, job.corp, job.condition, job.location, job.tag, job.link})
			done <- true
		}(job)
	}
	for i := 0; i < len(jobs); i++ {
		<-done
	}
}

func getPage(page int, mainC chan<- []extractedJob) {

	var jobs [] extractedJob
	c := make(chan extractedJob)

	pageURL := "https://www.saramin.co.kr/zf_user/search/recruit?searchword="+ keyword +"&go=&flag=n&searchMode=1&searchType=search&search_done=y&search_optional_item=n&recruitPage=" + strconv.Itoa(page) + "&recruitSort=relation&recruitPageCount=100&inner_com_type=&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&show_applied=&quick_apply=&except_read=&ai_head_hunting=&mainSearch=n"
	// fmt.Println("Requesting", pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit")

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})

	for i:=0; i<searchCards.Length(); i++ {
		job := <- c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	title := card.Find(".job_tit>a").Text()
	corp := cleanStirng(card.Find(".corp_name").Text())
	condition := cleanStirng(card.Find(".job_condition>span:nth-child(n+2)").Text())
	location := cleanStirng(card.Find(".job_condition>span>a").Text())
	link, _ := card.Attr("value")
	link = "https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=" + cleanStirng(link)
	tag := cleanStirng(card.Find(".job_sector").Text())

	c <- extractedJob{
		title: title,
		corp: corp,
		condition: condition,
		location: location,
		tag: tag,
		link: link,
	}  
}

func getPages() int {
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	// extaract search count
	slice := strings.Split(doc.Find(".cnt_result").Text(), " ")
	page_str := strings.Trim(slice[1], "건")
	slice = strings.Split(page_str, ",")
	page_str = strings.Join(slice, "")
	
	result, err := strconv.Atoi(page_str);
	checkErr(err)
	// fmt.Println(result)

	pages := result/100 + 1
	// fmt.Println("pages : ", pages)

	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

func cleanStirng(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}