package main

import (
	"strings"
	"errors"
	"strconv"
	"fmt"
	"io/ioutil"
	"regexp"
	"os"
)

func parseLine(in string) (string, error) {
	var out string

	var phrases = regexp.MustCompile("[.,:]+").Split(in, -1)

	for _, phrase := range phrases {
		phrase = strings.Trim(phrase, " \t")
		arr := strings.SplitN(phrase, " ", 2)
		if len(arr) == 0 {
			continue
		}
		if len(arr) == 1 {
			arr = append(arr, "")
		}
		switch strings.ToLower(arr[0]) {
			case "say", "saying", "talk", "talking", "write", "writing":			
				word := regexp.MustCompile("\".*\"").FindString(arr[1])
				word = word[1:len(word) - 1]
				word = strings.Replace(word, "\\", "\\\\", -1)
				word = strings.Replace(word, "\"", "\\\"",  -1)
				out += "puts(\"" + word + "\");\n"
			case "sleep", "sleeping", "wait", "waiting":
				str := regexp.MustCompile("[0-9]+").FindString(arr[1])
				sec, err := strconv.Atoi(str)
				if err == nil {
					out += "sleep(" + strconv.Itoa(sec * 1000) + ");\n"
				}
			case "failing", "fail":
				str := regexp.MustCompile("[0-9]+").FindString(arr[1])
				status, err := strconv.Atoi(str)
				if err != nil || status == 0 {
					status = 1
				}
				out += "exit(" + strconv.Itoa(status) + ");\n"
		}
	}

	return out, nil
}

// Converts essaylang into c
func Convert(in string) (string, error) {
	// TODO: Double spacing; every other line must be empty
	lines := strings.Split(in, "\n")
	if len(lines) < 6 {
		return "", errors.New(strconv.Itoa(len(lines)) + ": Missing " + (map[int]string{
			0: "name",
			1: "teacher",
			2: "subject",
			3: "date",
			4: "title",
			5: "essay",	
		}[len(lines)]))
	}
	var name, teacher, subject, date, title string
	var out string = `
#include <stdio.h>

int main() {
`
	for i, line := range lines {
		if len(line) == 0 {
			return "", errors.New(strconv.Itoa(i + 1) + ": Extraneous")
		}
		switch i {
		case 0:
			name = line
		case 1:
			teacher = line
		case 2:
			subject = line
		case 3:
			date = line
		case 4:
			if line[0] != byte('\t') || line[1] != byte('\t') {
				return "", errors.New("5: Title cannot be left-justified")
			}
			if regexp.MustCompile("(dying|die|killing|kill|murder|stab|stabbing)").FindStringIndex(line) != nil {
				return "", errors.New("5: Violence is not okay")
			}
			title = line
		default:
			if line[0] != byte('\t') {
				return "", errors.New(strconv.Itoa(i + 1) + ": Paragraph not indented")
			}
			newline, err := parseLine(line[1:])
			if err != nil {
				return "", errors.New(strconv.Itoa(i + 1) + ": " + err.Error())
			}
			out += newline
		}
	}

	_, _, _, _, _ =  name, teacher, subject, date, title
	
	out += "\treturn 0;\n}\n"	
	return out, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s file.txt\n", os.Args[0])
		return
	}

	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	out, err := Convert(string(content))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(out)
}