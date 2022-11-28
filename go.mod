module github.com/bootjp/jenkins_consecutive_fail_detector

go 1.19

require (
	github.com/ashwanthkumar/slack-go-webhook v0.0.0-20200209025033-430dd4e66960
	github.com/bndr/gojenkins v1.1.0
	github.com/elazarl/goproxy v0.0.0-20201021153353-00ad82a08272 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/tools v0.3.0
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/bndr/gojenkins => github.com/bootjp/gojenkins v1.0.2
