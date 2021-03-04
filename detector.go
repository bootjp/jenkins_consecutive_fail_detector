package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/bndr/gojenkins"
	"github.com/pkg/errors"
)

const JenkinsResultFail = "FAILURE"
const JenkinsResultSuccess = "SUCCESS"
const JenkinsResultAbort = "ABORTED"
const JenkinsResultRunning = "RUNNING"

const ReasonOverHoursFailedJob = "ErrOverHoursFailedJob"
const ReasonConsecutiveFailJob = "ConsecutiveFailJob"
const ReasonJenkinsError = "jenkins error"

var logger *log.Logger

func main() {

	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime)

	url := flag.String("url", "", "jenkins server url")
	flag.Parse()
	if *url == "" {
		log.Fatalln("jenkins url is required. -url https://example.com:8080/jenkins ")
	}

	fmt.Printf("jenkins url: %s\n", *url)

	jenkinsToken := os.Getenv("JENKINS_TOKEN")
	jenkinsUser := os.Getenv("JENKINS_USER")
	jenkinsPassword := os.Getenv("JENKINS_PASSWORD")
	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	slackUsername := os.Getenv("SLACK_USERNAME")

	// Versions lower than v0.0.5 have incorrect settings,
	// so load with typo for compatibility

	slackChannel := os.Getenv("SLACK_CHANNEL")
	if slackChannel == "" {
		slackChannel = os.Getenv("SLACK_CHANNNEL")
	}

	var jenkins *gojenkins.Jenkins
	if jenkinsToken != "" {
		jenkins = JenkinsInit(*url, jenkinsUser, jenkinsToken)
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

	var errs string
	for reason, failJobs := range summary {

		switch reason {
		case ReasonJenkinsError:
			errs += fmt.Sprintln("Jobs whose status could not be confirmed due to a Jenkins error")
		case ReasonOverHoursFailedJob:
			errs += fmt.Sprintln("Jobs that have been failed for over an hour")
		case ReasonConsecutiveFailJob:
			errs += fmt.Sprintln("Jobs that have failed more than once in a row")
		}

		for _, failJob := range failJobs {
			if failJob.Err != nil {
				errs += fmt.Sprintln(failJob.Err)
			}

			errs += fmt.Sprintln(failJob.JenkinsJob.GetName())

			if lfb, err := failJob.JenkinsJob.GetLastFailedBuild(); err == nil && lfb != nil {
				errs += fmt.Sprintln(lfb.GetUrl())
			}
		}

		errs += fmt.Sprintln("---")
	}

	fmt.Println(errs)

	if slackWebhookURL != "" && len(detectFailJobs) > 0 {
		payload := slack.Payload{
			Text:      errs,
			Username:  "jenkins_consecutive_fail_detector",
			IconEmoji: ":warning:",
		}

		if slackUsername != "" {
			payload.Username = slackUsername
		}
		if slackChannel != "" {
			payload.Channel = slackChannel
		}

		err := slack.Send(slackWebhookURL, "", payload)
		if len(err) > 0 {
			fmt.Printf("webhook error: %s\n", err)
		}
	}

	if len(detectFailJobs) > 0 {
		os.Exit(1)
	}

	fmt.Printf("%d jonbs checked. all status success or retring now\n", len(jobs))
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
			ids, _ := job.GetAllBuildIds()
			fmt.Println(ids)

			logger.Println("got err GetLastBuild by " + job.GetName() + " call by DetectFailJobs")
			fmt.Printf("%v", lastBuild)
			logger.Println(err)
			logger.Println(errors.WithStack(err))

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
	fmt.Println(job.GetAllBuildIds())
	latestBuild, err := job.GetLastBuild()
	if err != nil {
		logger.Println("got err GetLastBuild by " + job.GetName())
		fmt.Printf("%v", latestBuild)
		logger.Println(err)
		logger.Println(errors.WithStack(err))

		return false, err
	}

	return time.Since(latestBuild.GetTimestamp()) > 1*time.Hour, nil
}

func IsConsecutiveFailJob(job *gojenkins.Job) (bool, error) {
	buildIds, err := job.GetAllBuildIds()

	if err != nil {
		logger.Println("got err GetAllBuild by " + job.GetName() + "")
		logger.Println(err)
		logger.Println(errors.WithStack(err))

		return false, err
	}

	if len(buildIds) > 0 {
		return false, nil
	}

	fmt.Println(buildIds)
	fmt.Println(job.GetName())

	lastBuild, err := job.GetLastBuild()
	if err != nil {
		logger.Println("got err GetLastBuild by " + job.GetName())
		fmt.Printf("%v", lastBuild)
		logger.Println(err)
		logger.Println(errors.WithStack(err))

		return false, err
	}

	for _, buildId := range buildIds {
		if buildId.Number == lastBuild.GetBuildNumber() {
			// ignore last build.
			continue
		}

		build, err := job.GetBuild(buildId.Number)
		if err != nil {
			logger.Println("got err GetBuild")
			logger.Println(err)
			logger.Println(errors.WithStack(err))

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
