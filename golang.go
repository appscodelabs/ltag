package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type golangApplier struct {
}

func (g *golangApplier) CheckHeader(target *os.File, t *TagContext) (bool, error) {

	//Check compiler flags.
	cFlags, cbuf, err := g.checkSpecialConditions(target)
	if err != nil {
		return false, err
	}
	target.Seek(0, 0)

	tbuf, err := ioutil.ReadFile(filepath.Join(t.templatePath, "go.txt"))
	if err != nil {
		return false, err
	}

	if cFlags == AutoGenerated {
		return true, nil
	}

	var templateBuf string
	if cFlags == CompilerFlags {
		templateBuf = fmt.Sprintf("%s%s%s", cbuf, "\n\n", tbuf)
	} else {
		templateBuf = string(tbuf)
	}

	targetBuf := make([]byte, len(templateBuf))

	n, err := target.Read(targetBuf)
	if err != nil {
		return false, err
	}

	if n == len(templateBuf) {
		if strings.Compare(string(templateBuf), string(targetBuf)) == 0 {
			return true, nil
		}
		// Test for existing non-matching Copyright notice
		re := regexp.MustCompile(`/*(\s)*Copyright `)
		if re.FindStringIndex(string(targetBuf)) != nil {
			return true, nil
		}
	}

	return false, nil
}

func (g *golangApplier) ApplyHeader(path string, t *TagContext) error {

	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	t.templateFiles.goTemplateFile.Seek(0, 0)

	headerExist, err := g.CheckHeader(file, t)
	if err != nil {
		return err
	}

	if headerExist {
		return nil
	}

	//Reset the read pointers to beginning of file.
	t.templateFiles.goTemplateFile.Seek(0, 0)
	file.Seek(0, 0)

	sFlags, flags, err := g.checkSpecialConditions(file)
	if err != nil {
		return err
	}
	file.Seek(0, 0)

	tempFile := path + ".tmp"
	tFile, err := os.OpenFile(tempFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer tFile.Close()

	reader := bufio.NewReader(file)
	if sFlags == CompilerFlags {
		tFile.Write(flags)
		tFile.Write([]byte("\n\n"))
		_, _, err = reader.ReadLine()
		_, _, err = reader.ReadLine()
	}

	if sFlags == AutoGenerated {
		//This should not hit.
		panic(err)
	}

	_, err = io.Copy(tFile, t.templateFiles.goTemplateFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(tFile, reader)
	if err != nil {
		return err
	}

	err = os.Rename(tempFile, path)
	if err != nil {
		return err
	}
	return nil
}

func (g *golangApplier) checkSpecialConditions(target *os.File) (uint8, []byte, error) {
	reader := bufio.NewReader(target)
	buf, _, err := reader.ReadLine()
	if err != nil {
		return NormalFiles, nil, err
	}

	// checks for Package comments as per https://blog.golang.org/godoc-documenting-go-code
	if strings.HasPrefix(string(buf), "//") &&
		(strings.Contains(string(buf), "build") ||
			strings.Contains(string(buf), "unix") ||
			strings.Contains(string(buf), "linux") ||
			strings.Contains(string(buf), "windows") ||
			strings.Contains(string(buf), "darwin") ||
			strings.Contains(string(buf), "freebsd")) &&
		!strings.Contains(string(buf), "Package") {
		return CompilerFlags, buf, nil
	}
	if strings.HasPrefix(string(buf), "//") &&
		(strings.Contains(string(buf), "DO NOT EDIT")) {
		return AutoGenerated, nil, nil
	}
	return NormalFiles, nil, nil
}

func (g *golangApplier) RemoveHeader(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	start, end := FindGoCopyright(lines)
	if start > -1 && end > -1 {
		lines = append(lines[:start], lines[end+1:]...)
	}

	return ioutil.WriteFile(filename, TrimBlanks(lines), 0644)
}
