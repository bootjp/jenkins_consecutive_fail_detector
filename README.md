# jenkins_consecutive_fail_detector

Notify when there is a consecutive of fail in a job that fails due to various factors.

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
