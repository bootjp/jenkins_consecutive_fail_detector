<a href="https://codeclimate.com/github/bootjp/jenkins_consecutive_fail_detector/maintainability"><img src="https://api.codeclimate.com/v1/badges/73689e32f0fd15762eb6/maintainability" /></a>

# jenkins_consecutive_fail_detector

Modern systems and jobs are complex, and jobs can fail due to various system reasons.
However, it's a bit noisy to get a failure notification every time a job can be retried.
This tool ignores the job being retried and only notifies you when certain conditions are met.

The conditions for notification are

- More than n consecutive errors occur in a specific job.
- There is no job running even if n hours have passed since the failed build of the specific job.


## How to use

### install with go get

```bash
go get github.com/bootjp/jenkins_consecutive_fail_detector
JENKINS_TOKEN="secret_token" jenkins_consecutive_fail_detector -url http://example.com:8080/jenkins 
# or
JENKINS_USER="login_user" JENKINS_PASSWORD="login_password" jenkins_consecutive_fail_detector -url http://example.com:8080/jenkins 
```

### install release binary
```bash
curl -LO https://github.com/bootjp/jenkins_consecutive_fail_detector/releases/download/v0.0.0/jenkins_consecutive_fail_detector-linux-amd64
chmod +x jenkins_consecutive_fail_detector-linux-amd64
JENKINS_TOKEN="secret_token" jenkins_consecutive_fail_detector -url http://example.com:8080/jenkins
# or
JENKINS_USER="login_user" JENKINS_PASSWORD="login_password" jenkins_consecutive_fail_detector -url http://example.com:8080/jenkins 
```

### Slack notification

- Add environment value `SLACK_WEBHOOK` are enable slack webhook.
- Add environment value `SLACK_USERNAME` modify notify slack username.
- Add environment value `SLACK_CHANNNEL` modify notify slack channel.
