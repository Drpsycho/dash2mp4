package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type ByLength []string

// regexp for detect files and sort it
// example name file: samefilename_X_XXX.mp4
var digitsRegexp = regexp.MustCompile(`(\w+?)_(\d)_(\d+)\.mp4`)

// regexp for detect first file
// example: samefilename_X_.mp4
var firstfile = regexp.MustCompile(`\w+_\d_\.mp4`)

func (s ByLength) Len() int {
	return len(s)
}

func (s ByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLength) Less(i, j int) bool {

	res_i := digitsRegexp.FindStringSubmatch(s[i])
	res_j := digitsRegexp.FindStringSubmatch(s[j])

	if len(res_i) == 0 {
		return false
	}
	if len(res_j) == 0 {
		return false
	}
	res_i_in_digit, _ := strconv.Atoi(res_i[3])
	res_j_in_digit, _ := strconv.Atoi(res_j[3])

	return res_i_in_digit < res_j_in_digit
}

func clearfile(filename string) {
	os.Remove(filename)
}

func packmp4(file string, outputfilename string) {

	fmt.Println(file)
	r, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	w, err := os.OpenFile(outputfilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		panic(err)
	}
	defer w.Close()

	// do the actual work
	n, err := io.Copy(w, r)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Copied %v bytes\n", n)
}

func findfiles(files []os.FileInfo, outputfilename string, inputname string) {
	var filelist ByLength

	for _, file := range files {
		if !(file.Mode().IsRegular()) {
			continue
		}

		if !(filepath.Ext(file.Name()) == ".mp4") {
			continue
		}

		if file.Name() == outputfilename {
			continue
		}

		matched, _ := filepath.Match(inputname+"*", file.Name())
		if !(matched) {
			continue
		}

		res := firstfile.MatchString(file.Name())
		if res {
			packmp4(file.Name(), outputfilename)
		} else {
			filelist = append(filelist, file.Name())
		}

	}

	sort.Sort(filelist)

	for _, filename := range filelist {
		packmp4(filename, outputfilename)
	}
}

func main() {

	outputname := flag.String("o", "out.mp4", "Name for output file")
	pathtofolder := flag.String("p", "."+string(filepath.Separator), "Path to work folder")
	inputname := flag.String("i", "All", "Base name for input files")
	flag.Parse()

	if *inputname == "All" {
		*inputname = ""
	}

	d, err := os.Open(*pathtofolder)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	clearfile(*outputname)
	findfiles(files, *outputname, *inputname)
}
