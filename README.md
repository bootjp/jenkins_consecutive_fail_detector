# jenkins_consecutive_fail_detector

Notify when there is a consecutive of fail in a job that fails due to various factors.

The conditions for notification are

- More than n consecutive errors occur in a specific job.
- There is no job running even if n hours have passed since the failed build of the specific job.
