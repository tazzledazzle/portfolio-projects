package main

import (
	"encoding/xml"
	"io"
	"strconv"
	"time"
)

type TestResult struct {
	Name     string
	Status   string
	Duration time.Duration
}

type TestSuite struct {
	Name  string
	Tests []TestResult
}

type junitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	TestSuites []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure"`
	Skipped   *junitSkipped `xml:"skipped"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Content string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr"`
}

func ParseJUnit(r io.Reader) ([]TestResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []TestResult{}, nil
	}

	var results []TestResult

	var suites junitTestSuites
	if err := xml.Unmarshal(data, &suites); err == nil && len(suites.TestSuites) > 0 {
		for _, suite := range suites.TestSuites {
			results = append(results, parseTestCases(suite.TestCases)...)
		}
		return results, nil
	}

	var suite junitTestSuite
	if err := xml.Unmarshal(data, &suite); err != nil {
		return nil, err
	}

	return parseTestCases(suite.TestCases), nil
}

func parseTestCases(cases []junitTestCase) []TestResult {
	results := make([]TestResult, 0, len(cases))

	for _, tc := range cases {
		result := TestResult{
			Name:   tc.Name,
			Status: "passed",
		}

		if tc.Failure != nil {
			result.Status = "failed"
		} else if tc.Skipped != nil {
			result.Status = "skipped"
		}

		if tc.Time != "" {
			if secs, err := strconv.ParseFloat(tc.Time, 64); err == nil {
				result.Duration = time.Duration(secs * float64(time.Second))
			}
		}

		results = append(results, result)
	}

	return results
}
