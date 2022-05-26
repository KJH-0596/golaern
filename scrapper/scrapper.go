package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id			string
	title		string
	location	string
	salary		string
	summary		string
}


// Scrape indeed by term
func Scrape(term string){
	var baseURL string = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)

	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, c)
	}

	for i := 0; i < totalPages; i++ {
		extracteJobs := <-c
		jobs = append(jobs, extracteJobs...)
	}

	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, url string, mainC chan<- []extractedJob){
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := url + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extracteJob(card, c)
	})
	
	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

func extracteJob(card *goquery.Selection, c chan <- extractedJob) {
	id, _ := card.Find("h2>a").Attr("data-jk")
	// id = id
	title := CleanString(card.Find("h2>a>span").Text())
	location := CleanString(card.Find(".companyLocation").Text())
	salary := CleanString(card.Find(".salary-snippet-container").Text())
	summary := CleanString(card.Find(".job-snippet").Text())
	c <- extractedJob{
		id: id,
		title: title,
		location: location,
		salary: salary,
		summary: summary,
	}
}

// getPages return number of url's pages
func getPages(url string) int{
	pages := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func writeJobs(jobs []extractedJob){
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link", "title", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/채용보기?jk=" + job.id, job.title, job.location, job.salary, job.summary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response){
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}


func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}


// func blog() {
// 	file1, _ := ioutil.ReadFile("addr2.txt")
// 	var urls []string = strings.Split(string(file1), "\n")
	
// 	for index, url := range urls{
// 		blogURL := url
// 		res, err := http.Get(blogURL)
// 		checkErr(err)
// 		checkCode(res)

// 		defer res.Body.Close()

// 		doc, err := goquery.NewDocumentFromReader(res.Body)
// 		checkErr(err)

// 		ext := doc.Find(".se-text-paragraph")
// 		result := CleanString(ext.Text()) + "\n"

// 		result = strings.Replace(result, ".", "\n", -1)
// 		fmt.Println(index+1,url, "...Done!")
// 		fmt.Println(result)
// 	}
// }

