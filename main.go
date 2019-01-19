package main

import (
	"strings"
	"errors"
	"strconv"
	"fmt"
	"io/ioutil"
	"regexp"
	"os"
	"flag"
	"os/exec"
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
					out += "sleep(" + strconv.Itoa(sec) + ");\n"
				}
			case "failing", "fail":
				str := regexp.MustCompile("[0-9]+").FindString(arr[1])
				status, err := strconv.Atoi(str)
				if err != nil || status == 0 {
					status = 1
				}
				out += "exit(" + strconv.Itoa(status) + ");\n"
			case "stop":
				out += "exit(0);\n"
			case "read":
				out += "getchar();\n"
		}
	}

	return out, nil
}

// Converts essaylang into c
func Convert(in string) (string, error) {
	// TODO: Double spacing; every other line must be empty	
	lines := strings.Split(in, "\n")
	
	if len(lines) < 13 {
		return "", errors.New(strconv.Itoa(len(lines)) + ": Missing " + (map[int]string{
			0: "name",
			1: "teacher",
			2: "subject",
			3: "date",
			4: "title",
			5: "essay",	
		}[(len(lines) + 1) / 2]))
	}
	var name, teacher, subject, date, title string
	var out string = `
#include <stdio.h>
#include <unistd.h>

int main() {
`
	for i, line := range lines {
		if i % 2 == 1 {
			if len(line) != 0 {
				return "", errors.New(strconv.Itoa(i + 1) + ": Essays must be double spaced")
			}
			continue
		}
	
		if len(line) == 0 {
			return "", errors.New(strconv.Itoa(i + 1) + ": Extraneous")
		}
		switch i {
		case 0:
			name = line
		case 2:
			teacher = line
		case 4:
			subject = line
		case 6:
			date = line
		case 8:
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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-c cc] file.txt\n", os.Args[0])		
	}

	var compile string
	flag.StringVar(&compile, "c", "", "Compiler to use, or none to not compile")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
		return
	}

	content, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: " + err.Error())
		os.Exit(1)
		return
	}
	out, err := Convert(string(content))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: " + err.Error())
		os.Exit(1)		
		return
	}

	if len(compile) == 0 {
		fmt.Println(out)
	} else {
		path := os.TempDir() + "/essay.c"
		tmp, err := os.Create(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: " + err.Error())
		}
		tmp.Write([]byte(out))
		tmp.Close()
		
		exec.Command(compile, path).Run()
	}
}