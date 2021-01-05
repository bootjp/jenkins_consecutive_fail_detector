package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bndr/gojenkins"
)

const ResultFail = "FAILURE"
const ResultSuccess = "SUCCESS"
const ResultAbort = "ABORTED"
const ResultRunning = "RUNNING"

func main() {

	url := flag.String("url", "", "jenkins server url")
	flag.Parse()
	if *url == "" {
		log.Fatalln("jenkins url is required. -url http://example.com:8080/jenkins ")
	}

	fmt.Printf("jenkins url: %s\n", *url)

	jenkinsToken := os.Getenv("JENKINS_TOKEN")
	jenkinsUser := os.Getenv("JENKINS_USER")
	jenkinsPassword := os.Getenv("JENKINS_PASSWORD")

	var jenkins *gojenkins.Jenkins
	if jenkinsToken != "" {
		jenkins = JenkinsInit(*url, jenkinsToken)
	} else {
		jenkins = JenkinsInit(*url, jenkinsUser, jenkinsPassword)
	}

	j, err := jenkins.Init()
	if err != nil {
		log.Fatalln(err)
	}

	jobs, err := j.GetAllJobs()
	if err != nil {
		log.Fatalln(err)
	}

	detectFailJobs := DetectFailJobs(jobs)

	for _, job := range detectFailJobs {
		fmt.Println("----")
		fmt.Println(job.Reason)
		fmt.Println(job.Job.GetName())
		if job.Err != nil {
			fmt.Println(job.Err)
		}
	}

	exitCode := 0
	if len(detectFailJobs) > 0 {
		exitCode = 1
	}
	os.Exit(exitCode)
}

func JenkinsInit(url string, auth ...string) *gojenkins.Jenkins {
	if len(auth) == 2 {
		return gojenkins.CreateJenkins(
			nil,
			url,
			auth[0],
			auth[1],
		)
	}

	return gojenkins.CreateJenkins(
		nil,
		url,
		auth[0],
	)

}

type FailJob struct {
	Job    *gojenkins.Job
	Err    error
	Reason string
}

func DetectFailJobs(jobs []*gojenkins.Job) []*FailJob {

	var errorJobs []*FailJob
	for _, job := range jobs {
		enable, err := job.IsEnabled()
		if err != nil {
			log.Fatalln(err)
		}
		if !enable {
			continue
		}

		lastBuild, err := job.GetLastBuild()
		if err != nil {

			ej := &FailJob{
				Job:    job,
				Err:    err,
				Reason: "error",
			}
			errorJobs = append(errorJobs, ej)
			continue
		}

		switch lastBuild.GetResult() {
		case ResultFail:

			fail, err := IsOverHoursFailedJob(job)
			if err != nil {
				ej := &FailJob{
					Job:    job,
					Err:    err,
					Reason: "error",
				}

				errorJobs = append(errorJobs, ej)
			}

			if fail {
				ej := &FailJob{
					Job:    job,
					Err:    err,
					Reason: "IsOverHoursFailedJob",
				}

				errorJobs = append(errorJobs, ej)
				continue
			}

			fail, err = IsConsecutiveFailJob(job)

			if err != nil {
				ej := &FailJob{
					Job:    job,
					Err:    err,
					Reason: "error",
				}
				errorJobs = append(errorJobs, ej)
			}

			if fail {
				ej := &FailJob{
					Job:    job,
					Err:    err,
					Reason: "IsConsecutiveFailJob",
				}
				errorJobs = append(errorJobs, ej)
				continue
			}

		case ResultRunning:
			// retrying job is ignore
			continue
		}
	}

	return errorJobs
}

func IsOverHoursFailedJob(job *gojenkins.Job) (bool, error) {
	latestBuild, err := job.GetLastBuild()
	if err != nil {
		return false, err
	}

	if time.Since(latestBuild.GetTimestamp()) > 1*time.Hour {
		return true, nil
	}
	return false, nil
}

func IsConsecutiveFailJob(job *gojenkins.Job) (bool, error) {
	buildIds, err := job.GetAllBuildIds()

	if err != nil {
		return false, err
	}

	lastBuild, err := job.GetLastBuild()
	if err != nil {
		return false, err
	}

	for _, buildId := range buildIds {
		if buildId.Number == lastBuild.GetBuildNumber() {
			// ignore last build.
			continue
		}

		build, err := job.GetBuild(buildId.Number)
		if err != nil {
			return false, err
		}

		switch build.GetResult() {
		case ResultFail:
			return true, nil

		case ResultAbort:
			continue
		case ResultSuccess:
			return false, nil
		}

		break
	}

	return false, nil
}
