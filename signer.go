package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)



var ExecutePipeline = func(jobs ...job) {
	//Create workGroup
	wg := &sync.WaitGroup{}
	//Open in chanell
	in := make(chan interface{})

	//for Distribution of work
	for _, job := range jobs {
		wg.Add(1)

		out := make(chan interface{})
		go startWorker(job, in, out, wg)

		in = out
	}
	//Waiting for the end
	wg.Wait()
}

//func Distribution of work
func startWorker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)

	job(in, out)
}

var SingleHash = func(in, out chan interface{}) {
	//Create workGroup and mutex
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)

		go func(in interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			//int to string
			data := strconv.Itoa(in.(int))

			fmt.Printf("%s SingleHash data %s\n", data, data)
			//Lock function for 1 guartin
			mu.Lock()
			md5Data := DataSignerMd5(data)
			mu.Unlock()
			fmt.Printf("%s SingleHash md5(data) %s\n", data, md5Data)

			crc32DataChan := make(chan string)
				go func(data string, out chan string) {

					out <- DataSignerCrc32(data)

				}(data, crc32DataChan)

			crc32Md5Data := DataSignerCrc32(md5Data)
			fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32Md5Data)

			crc32Data := <-crc32DataChan
			fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32Data)
			//Sending the result
			out <- crc32Data + "~" + crc32Md5Data
			fmt.Printf("%s SingleHash result %s\n", data, crc32Data+"~"+crc32Md5Data)
		}(i, out, wg, mu)
	}

	wg.Wait()
}


var MultiHash = func(in, out chan interface{}) {
	//Create workGroup
	wg := &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)


		go func(in string, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			wg2 := &sync.WaitGroup{}
			//Creating an empty array
			concatArray := make([]string, 6)
			for i := 0; i < 6; i++ {
				wg2.Add(1)
				//adding i to a string
				data := strconv.Itoa(i) + in
				go func(concatArray []string, data string, i int, wg *sync.WaitGroup) {
					defer wg.Done()

					data = DataSignerCrc32(data)
					//Writing to an array
					concatArray[i] = data
					fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", in, i, data)


				}(concatArray, data, i, wg2)
			}
			wg2.Wait()
			//concatenates array
			result := strings.Join(concatArray, "")
			fmt.Printf("%s MultiHash result: %s\n", in, result)

			out <- result
		}(i.(string), out, wg)
	}

	wg.Wait()
}


var CombineResults = func(in, out chan interface{}) {
	//Create array
	var arrayString []string
	//filling
	for i := range in {
		arrayString = append(arrayString, i.(string))
	}
	//sort,concatenates the elements to create a single string, print
	sort.Strings(arrayString)
	result := strings.Join(arrayString, "_")
	fmt.Printf("CombineResults \n%s\n", result)
	//transmission of results
	out <- result
}
