package test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/wallarm/gotestwaf/internal/data/config"
	"gopkg.in/yaml.v2"
)

const testCaseExt = ".yml"

func Load(cfg *config.Config, logger *log.Logger) ([]Case, error) {
	var files []string
	var testCases []Case

	if cfg.TestCasesPath == "" {
		return nil, errors.New("empty test cases path")
	}

	if err := filepath.Walk(cfg.TestCasesPath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, err
	}

	logger.Println("Loading test cases: ")
	for _, testCaseFile := range files {
		if filepath.Ext(testCaseFile) != testCaseExt {
			continue
		}

		// Ignore subdirectories, process as .../<testSetName>/<testCaseName>/<case>.yml
		parts := strings.Split(testCaseFile, "/")
		parts = parts[len(parts)-3:]

		testSetName := parts[1]
		testCaseName := strings.TrimSuffix(parts[2], testCaseExt)

		if cfg.TestSet != "" && testSetName != cfg.TestSet {
			continue
		}

		if cfg.TestCase != "" && testCaseName != cfg.TestCase {
			continue
		}

		logger.Printf("%v:%v", testSetName, testCaseName)

		yamlFile, err := ioutil.ReadFile(testCaseFile)
		if err != nil {
			return nil, err
		}

		var t Case
		err = yaml.Unmarshal(yamlFile, &t)
		if err != nil {
			return nil, err
		}

		t.Name = testCaseName
		t.Set = testSetName

		if strings.Contains(testSetName, "false") {
			t.IsTruePositive = false // test case is false positive
		} else {
			t.IsTruePositive = true // test case is true positive
		}

		testCases = append(testCases, t)
	}

	if testCases == nil {
		return nil, errors.New("no tests were selected")
	}

	return testCases, nil
}