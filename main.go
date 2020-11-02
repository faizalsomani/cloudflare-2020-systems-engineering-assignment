package main

import (
	"fmt"
	"os"
	"strconv"
	"net"
	"net/url"
	"crypto/tls"
	"time"
	"io/ioutil"
	"sort"
	"sync"
)

// this function gives the position of an element in array
func find(element string, data []string) (int) {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

// this function exits if with msg if error occurs
func handleErr(e error, msg string){
	if e != nil {
		fmt.Println(msg)
		os.Exit(0)
	}
}

func getRequest(URL string, codeArr *[]int, timeArr *[]int, sizeArr *[]int, respArr *string, wg *sync.WaitGroup) {
	defer wg.Done()
	// create variables
	parsedURL, err := url.Parse(URL)
	handleErr(err, "Invalid URL")
	dialer := net.Dialer{
		Timeout: time.Minute,
	}
	startTime := time.Now()
	conn, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("%s:https", parsedURL.Host), nil)
	handleErr(err, "Connection error")
	defer conn.Close()

	// get result for URL
	conn.Write([]byte("GET " + parsedURL.Path + " HTTP/1.0\r\nHost: " + parsedURL.Host + "\r\n\r\n"))
	resp, err := ioutil.ReadAll(conn)
	handleErr(err, "Unable to read respose")

	endTime := time.Now()
	strResp := string(resp)
	respCode, _ := strconv.Atoi(strResp[9:12])
	
	// update variables passed by refrence
	if respCode >= 400 || respCode < 200 {
		*codeArr = append(*codeArr, respCode)
	}
	*timeArr = append(*timeArr, int(endTime.Sub(startTime).Milliseconds()))
	*sizeArr = append(*sizeArr, len(strResp))
	*respArr = strResp
}

func handleHelp() {
	fmt.Println("Usage:")
	fmt.Println("go run main --help")
	fmt.Println("go run main --url <URL>")
	fmt.Println("go run main --url <URL> --profile <Total Requests>")
}

func handleProfile(url string, profile int){
	// create array of stats
	codeArr := make([]int, 0)
	timeArr := make([]int, 0)
	sizeArr := make([]int, 0)
	respArr := ""

	// call getRequest inside waitegroup
	var wg sync.WaitGroup
	for i:=0; i < profile; i++ {
		wg.Add(1)
		go getRequest(url, &codeArr, &timeArr, &sizeArr, &respArr, &wg)
	}
	wg.Wait()

	// prepare stats
	sort.Ints(timeArr)
	sort.Ints(sizeArr)
	timeSum := 0
	for i:=0; i < len(timeArr); i++ {
		timeSum += timeArr[i]
	}
	medianTime := 0
	if len(timeArr)%2 == 0 {
		medianTime = (timeArr[len(timeArr)/2] + timeArr[len(timeArr)/2 - 1])/2
	} else {
		medianTime = timeArr[(len(timeArr)-1)/2]
	}

	// print
	fmt.Println("\n--------------Start Profile---------------")
	fmt.Println()
	fmt.Printf("Number of requests: %d\n", profile)
	fmt.Printf("Fastest Time: %d\n", timeArr[0])
	fmt.Printf("Slowest Time: %d\n", timeArr[len(timeArr)-1])
	fmt.Printf("Mean Time: %d\n", timeSum/profile)
	fmt.Printf("Median Time: %d\n", medianTime)
	fmt.Printf("Percent of successful requests: %.f%%\n", float64((profile-len(codeArr))/profile)*100)
	fmt.Printf("Error Codes: %v\n", codeArr)
	fmt.Printf("Size in bytes of the smallest response: %d\n", sizeArr[0])
	fmt.Printf("Size in bytes of the largest response: %d\n", sizeArr[len(sizeArr)-1])
	fmt.Println("\n---------------End Profile----------------")
	fmt.Println()
}

func handleURL(url string) {
	// create array of stats
	codeArr := make([]int, 0)
	timeArr := make([]int, 0)
	sizeArr := make([]int, 0)
	respArr := ""

	// call getRequest inside waitegroup
	var wg sync.WaitGroup
	wg.Add(1)
	go getRequest(url, &codeArr, &timeArr, &sizeArr, &respArr, &wg)
	wg.Wait()

	// print
	fmt.Println("\n--------------Start Response---------------")
	fmt.Println()
	fmt.Println(respArr)
	fmt.Println("\n---------------End Response----------------")
	fmt.Println()
}

func main(){
	args := os.Args[1:]
	urlIndex := find("--url", args)
	profileIndex := find("--profile", args)
	
	if profileIndex != -1 && urlIndex != -1 {
		profile, e := strconv.Atoi(args[profileIndex+1])
		if e != nil || profile < 1 {
			fmt.Println("Invalid Profile")
			handleHelp()
			os.Exit(0)
		}
		handleProfile(args[urlIndex+1], profile)
	} else if urlIndex != -1 {
		handleURL(args[urlIndex+1])
	} else {
		handleHelp()
	}
}