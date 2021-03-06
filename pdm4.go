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

type InputParam struct {
	outputname, pathtofolder, inputname string
}

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

func RemoveOutputFile(filename string) {
	os.Remove(filename)
}

func PackToMP4(file string, outputfilename string) {

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

func OpenFolder(pathtofolder string) []os.FileInfo {
	d, err := os.Open(pathtofolder)
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
	return files
}

func FindFiles(param InputParam) (ByLength, string) {

	files := OpenFolder(param.pathtofolder)

	var filelist ByLength
	var firstfilename string
	for _, file := range files {
		if !(file.Mode().IsRegular()) {
			continue
		}

		if !(filepath.Ext(file.Name()) == ".mp4") {
			continue
		}

		if file.Name() == param.outputname {
			continue
		}

		matched, _ := filepath.Match(param.inputname+"*", file.Name())
		if !(matched) {
			continue
		}

		res := firstfile.MatchString(file.Name())
		if res {
			firstfilename = file.Name()
			// PackToMP4(file.Name(), param.outputname)
		} else {
			filelist = append(filelist, file.Name())
		}

	}
	return filelist, firstfilename
}

func ParseInputParam() InputParam {
	var param InputParam
	outputname := flag.String("o", "out.mp4", "Name for output file")
	pathtofolder := flag.String("p", "."+string(filepath.Separator), "Path to work folder")
	inputname := flag.String("i", "All", "Base name for input files")
	flag.Parse()

	param.outputname = *outputname
	param.pathtofolder = *pathtofolder
	param.inputname = *inputname

	fmt.Printf("\ninput param:\n output name - %s;\n work dir path - %s;\n input filter (by name) - %s\n",
		param.outputname,
		param.pathtofolder,
		param.inputname)

	if param.inputname == "All" {
		param.inputname = ""
	}

	RemoveOutputFile(param.outputname)

	return param
}

func ResultFileStat(name string) {
	stat, _ := os.Stat(name)
	fmt.Println("Create file ", name, " size - ", stat.Size())
}

func main() {
	param := ParseInputParam()
	filelist, firstfilename := FindFiles(param)
	fmt.Println("\nCopy files to ", param.outputname, ":")

	if firstfilename == "" {
		fmt.Println("error! no first file youstream_0_.mp4")
		os.Exit(1)
	}

	PackToMP4(firstfilename, param.outputname)
	sort.Sort(filelist)
	for _, filename := range filelist {
		if digitsRegexp.MatchString(filename) {
			PackToMP4(filename, param.outputname)
		}
	}
	ResultFileStat(param.outputname)
}
