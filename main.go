package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

func main() {
	runtime.GOMAXPROCS(4)
	start := time.Now()
	files := checkExt(".md")
	for j := 0; j < len(files); j++ {
		wg.Add(1)
		go workOnFile(files[j])
	}
	wg.Wait()
	elapsedTime := time.Since(start)
	fmt.Println("Total Time For Execution: " + elapsedTime.String())

}

func workOnFile(filestr string) {

	input, err := ioutil.ReadFile(filestr)
	check(err)

	lines := strings.Split(string(input), "\n")

	i := 0
	j := 0
	for lineindex, line := range lines {
		re1 := regexp.MustCompile(`!\[[^\)]*\]\(http\S*\)`)
		re2 := regexp.MustCompile(`\(.*\)`)
		re3 := regexp.MustCompile(`(\..{1,4}\))$`)
		res := re1.FindAllString(line, -1)
		if len(res) > 0 {
			var ext [2]string
			for ii, resstr := range res {
				rrr := strings.TrimPrefix(re2.FindString(resstr), "(")
				exttmp := re3.FindString(rrr)
				ext[ii] = strings.TrimSuffix(exttmp, ")")
				rrr = strings.TrimSuffix(rrr, ")")
				downloadFile("/Users/ShapedHorizon/wetravel/docsv2/images/"+filestr+"_img_"+strconv.Itoa(i)+ext[ii], rrr)
				i++
			}
			flag := false
			lines[lineindex] = re1.ReplaceAllStringFunc(lines[lineindex], func(a string) string {
				if flag {
					j++
					return "![images/" + filestr + "_img_" + strconv.Itoa(j) + "](/" + "images/" + filestr + "_img_" + strconv.Itoa(j) + ext[1] + ")"
				}
				flag = true
				return re1.ReplaceAllString(a, "![images/"+filestr+"_img_"+strconv.Itoa(j)+"](/"+"images/"+filestr+"_img_"+strconv.Itoa(j)+ext[0]+")")
			})
			j++
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filestr, []byte(output), 0666)
	check(err)
	wg.Done()
}

func downloadFile(filepath string, url string) {

	defer func() {
		if r := recover(); r != nil {
			log.Println(string(debug.Stack()))
			panic(errors.New("die"))
		}
	}()
	// defer wg.Done()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		check(err)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		check(err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func checkExt(ext string) []string {
	pathS, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var files []string
	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(path) == ext {
				files = append(files, f.Name())
			}
		}
		return nil
	})
	return files
}
