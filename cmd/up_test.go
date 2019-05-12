package cmd

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestUp_DoesntHangWhenCreationCancelled(t *testing.T) {
	if testing.Short() {
		t.Skip("skip e2e tests in short mode")
	}

	stackName := "test-cancelled-stack"
	RootCmd.SetArgs([]string{
		"up",
		"--always-succeed",
		"--stack-name", stackName,
		"--template", "../sample/sample.yml",
		"--param-value", "HealthCheckPath=/pinga",
	})

	buf := &bytes.Buffer{}
	out := io.MultiWriter(buf, os.Stderr)
	RootCmd.SetOutput(out)

	time.AfterFunc(5*time.Second, func() {
		for !strings.Contains(buf.String(), "LogGroup - CREATE_IN_PROGRESS") {
			time.Sleep(time.Second)
		}

		sess := session.Must(session.NewSession(aws.NewConfig().WithRegion("ap-southeast-2")))
		cfn := cloudformation.New(sess)
		input := &cloudformation.DeleteStackInput{StackName: &stackName}
		_, err := cfn.DeleteStack(input)
		if err != nil {
			panic(err)
		}
	})

	_ = RootCmd.Execute()

	assert.Regexp(t, regexp.MustCompile(`^\[\d\d:\d\d:\d\d] test-cancelled-stack - CREATE_IN_PROGRESS - User Initiated
\[\d\d:\d\d:\d\d]             LogGroup - CREATE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]             LogGroup - CREATE_IN_PROGRESS - Resource creation Initiated
\[\d\d:\d\d:\d\d]             LogGroup - CREATE_COMPLETE 
\[\d\d:\d\d:\d\d] test-cancelled-stack - DELETE_IN_PROGRESS - User Initiated
\[\d\d:\d\d:\d\d]             LogGroup - DELETE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]             LogGroup - DELETE_COMPLETE 
\[\d\d:\d\d:\d\d] test-cancelled-stack - DELETE_COMPLETE 
\{\}
`), buf.String())
}

func TestUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skip e2e tests in short mode")
	}

	t.Run("create", func(t *testing.T) {
		RootCmd.SetArgs([]string{
			"up",
			"--stack-name", "test-stack",
			"--template", "../sample/sample.yml",
			"--param-value", "HealthCheckPath=/pinga",
		})

		buf := &bytes.Buffer{}
		out := io.MultiWriter(buf, os.Stderr)
		RootCmd.SetOutput(out)

		_ = RootCmd.Execute()

		assert.Regexp(t, regexp.MustCompile(`^\[\d\d:\d\d:\d\d]           test-stack - CREATE_IN_PROGRESS - User Initiated
\[\d\d:\d\d:\d\d]             LogGroup - CREATE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]             LogGroup - CREATE_IN_PROGRESS - Resource creation Initiated
\[\d\d:\d\d:\d\d]             LogGroup - CREATE_COMPLETE 
\[\d\d:\d\d:\d\d]              TaskDef - CREATE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]              TaskDef - CREATE_IN_PROGRESS - Resource creation Initiated
\[\d\d:\d\d:\d\d]              TaskDef - CREATE_COMPLETE 
\[\d\d:\d\d:\d\d]          TargetGroup - CREATE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]          TargetGroup - CREATE_IN_PROGRESS - Resource creation Initiated
\[\d\d:\d\d:\d\d]          TargetGroup - CREATE_COMPLETE 
\[\d\d:\d\d:\d\d]           test-stack - CREATE_COMPLETE 
\{
  "LogGroup": "test-stack-LogGroup",
  "TaskDef": "arn:aws:ecs:ap-southeast-2:607481581596:task-definition/ecs-run-task-test:\d+"
\}
`), buf.String())
	})

	t.Run("update", func(t *testing.T) {
		RootCmd.SetArgs([]string{
			"up",
			"--stack-name", "test-stack",
			"--template", "../sample/sample.yml",
			"--param-value", "HealthCheckPath=/pingb",
		})

		buf := &bytes.Buffer{}
		out := io.MultiWriter(buf, os.Stderr)
		RootCmd.SetOutput(out)

		_ = RootCmd.Execute()

		assert.Regexp(t, regexp.MustCompile(`^\[\d\d:\d\d:\d\d]           test-stack - UPDATE_IN_PROGRESS - User Initiated
\[\d\d:\d\d:\d\d]          TargetGroup - UPDATE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]          TargetGroup - UPDATE_COMPLETE 
\[\d\d:\d\d:\d\d]           test-stack - UPDATE_COMPLETE_CLEANUP_IN_PROGRESS 
\[\d\d:\d\d:\d\d]           test-stack - UPDATE_COMPLETE 
\{
  "LogGroup": "test-stack-LogGroup",
  "TaskDef": "arn:aws:ecs:ap-southeast-2:607481581596:task-definition/ecs-run-task-test:\d+"
\}
`), buf.String())
	})

	t.Run("down", func(t *testing.T) {
		RootCmd.SetArgs([]string{
			"down",
			"--stack-name", "test-stack",
		})

		buf := &bytes.Buffer{}
		out := io.MultiWriter(buf, os.Stderr)
		RootCmd.SetOutput(out)

		_ = RootCmd.Execute()

		assert.Regexp(t, regexp.MustCompile(`^\[\d\d:\d\d:\d\d]           test-stack - DELETE_IN_PROGRESS - User Initiated
\[\d\d:\d\d:\d\d]          TargetGroup - DELETE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]          TargetGroup - DELETE_COMPLETE 
\[\d\d:\d\d:\d\d]              TaskDef - DELETE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]              TaskDef - DELETE_COMPLETE 
\[\d\d:\d\d:\d\d]             LogGroup - DELETE_IN_PROGRESS 
\[\d\d:\d\d:\d\d]             LogGroup - DELETE_COMPLETE`), buf.String())
	})

}
