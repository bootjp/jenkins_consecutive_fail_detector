package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestIsConsecutiveFailJob(t *testing.T) {

}

func TestIsOverHoursFailedJob(t *testing.T) {

}

func TestExcludeJobNamePattern(t *testing.T) {

	jobs := []string{"aa_bn", "ccc", "d", "eeee", "e"}

	re, err := regexp.Compile("e")
	if err != nil {
		t.Fatal(err)
	}

	filterd := make([]string, len(jobs))

	for _, job := range jobs {
		if !re.MatchString(job) {
			filterd = append(filterd, job)
			fmt.Println("un match", job)
		}
	}

	fmt.Println(filterd)

}
