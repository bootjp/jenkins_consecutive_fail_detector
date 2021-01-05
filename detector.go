package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bndr/gojenkins"
)

const JenkinsResultFail = "FAILURE"
const JenkinsResultSuccess = "SUCCESS"
const JenkinsResultAbort = "ABORTED"
const JenkinsResultRunning = "RUNNING"

const ReasonOverHoursFailedJob = "ErrOverHoursFailedJob"
const ReasonConsecutiveFailJob = "ConsecutiveFailJob"
const ReasonJenkinsError = "jenkins error"

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
	webhookURL := os.Getenv("WEBHOOK_URL")

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
	summary := map[string][]*FailJob{}
	for _, job := range detectFailJobs {
		summary[job.Reason] = append(summary[job.Reason], job)
	}

	for reason, failJobs := range summary {

		switch reason {
		case ReasonJenkinsError:
			fmt.Println("Jobs whose status could not be confirmed due to a Jenkins error")
		case ReasonOverHoursFailedJob:
			fmt.Println("Jobs that have been failed for over an hour")
		case ReasonConsecutiveFailJob:
			fmt.Println("Jobs that have failed more than once in a row")
		}

		for _, failJob := range failJobs {
			if failJob.Err != nil {
				fmt.Println(failJob.Err)
			}

			fmt.Println(failJob.JenkinsJob.GetName())

			if lfb, err := failJob.JenkinsJob.GetLastFailedBuild(); err == nil && lfb != nil {
				fmt.Println(lfb.GetUrl())
			}
		}

		fmt.Println("---")
	}

	if webhookURL != "" {
		// todo notify slack code
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
	JenkinsJob *gojenkins.Job
	Err        error
	Reason     string
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
				JenkinsJob: job,
				Err:        err,
				Reason:     ReasonJenkinsError,
			}
			errorJobs = append(errorJobs, ej)
			continue
		}

		switch lastBuild.GetResult() {
		case JenkinsResultFail:

			fail, err := IsOverHoursFailedJob(job)
			if err != nil {
				ej := &FailJob{
					JenkinsJob: job,
					Err:        err,
					Reason:     ReasonJenkinsError,
				}

				errorJobs = append(errorJobs, ej)
			}

			if fail {
				ej := &FailJob{
					JenkinsJob: job,
					Err:        err,
					Reason:     ReasonOverHoursFailedJob,
				}

				errorJobs = append(errorJobs, ej)
				continue
			}

			fail, err = IsConsecutiveFailJob(job)

			if err != nil {
				ej := &FailJob{
					JenkinsJob: job,
					Err:        err,
					Reason:     ReasonJenkinsError,
				}
				errorJobs = append(errorJobs, ej)
			}

			if fail {
				ej := &FailJob{
					JenkinsJob: job,
					Err:        err,
					Reason:     ReasonConsecutiveFailJob,
				}
				errorJobs = append(errorJobs, ej)
				continue
			}

		case JenkinsResultRunning:
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
		case JenkinsResultFail:
			return true, nil

		case JenkinsResultAbort:
			continue
		case JenkinsResultSuccess:
			return false, nil
		}

		break
	}

	return false, nil
}
