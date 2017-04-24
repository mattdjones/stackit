// Copyright © 2017 Aidan Steele <aidan.steele@glassechidna.com.au>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"io/ioutil"
	"strings"
	"github.com/glassechidna/stackit/stackit"
)

// up --stack-name stackit-test --template sample.yml --param-value DockerImage=nginx --param-value Cluster=app-cluster-Cluster-1C2I18JXK9QNM --tag MyTag=Cool

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring stack up to date",
	Run: func(cmd *cobra.Command, args []string) {
		region := viper.GetString("region")
		profile := viper.GetString("profile")
		stackName := viper.GetString("stack-name")

		serviceRole := viper.GetString("service-role")
		paramValues := viper.GetStringSlice("param-value")
		previousParamValues := viper.GetStringSlice("previous-param-value")
		tags := viper.GetStringSlice("tag")
		notificationArns := viper.GetStringSlice("notification-arn")
		stackPolicy := viper.GetString("stack-policy")
		template := viper.GetString("template")
		previousTemplate := viper.GetBool("previous-template")
		noDestroy := viper.GetBool("no-destroy")
		cancelOnExit := !viper.GetBool("no-cancel-on-exit")

		parsed := parseCLIInput(
			stackName,
			serviceRole,
			stackPolicy,
			template,
			paramValues,
			previousParamValues,
			tags,
			notificationArns,
			previousTemplate)

		stackit.Up(region, profile, parsed, noDestroy, cancelOnExit)
	},
}

func parseCLIInput(
	stackName,
	serviceRole,
	stackPolicy,
	template string,
	paramValues,
	previousParamValues,
	tags,
	notificationArns []string,
	previousTemplate bool) stackit.StackitUpInput {
	input := stackit.StackitUpInput{
		StackName: aws.String(stackName),
	}

	if len(serviceRole) > 0 {
		input.RoleARN = aws.String(serviceRole)
	}

	if len(stackPolicy) > 0 {
		policyBody, err := ioutil.ReadFile(stackPolicy)
		if err != nil {

		} else {
			input.StackPolicyBody = aws.String(string(policyBody))
		}
	}

	if len(template) > 0 {
		templateBody, err := ioutil.ReadFile(template)
		if err != nil {

		} else {
			input.TemplateBody = aws.String(string(templateBody))
		}
	}

	if previousTemplate {
		input.PreviousTemplate = aws.Bool(previousTemplate)
	}

	params := []*cloudformation.Parameter{}

	for _, paramPair := range paramValues {
		parts := strings.SplitN(paramPair, "=", 2)
		name, value := parts[0], parts[1]

		params = append(params, &cloudformation.Parameter{
			ParameterKey: aws.String(name),
			ParameterValue: aws.String(value),
		})
	}

	for _, param := range previousParamValues {
		params = append(params, &cloudformation.Parameter{
			ParameterKey: aws.String(param),
			UsePreviousValue: aws.Bool(true),
		})
	}

	input.Parameters = params

	if len(tags) > 0 {
		cfnTags := []*cloudformation.Tag{}

		for _, tagPair := range tags {
			parts := strings.SplitN(tagPair, "=", 2)
			name, value := parts[0], parts[1]

			cfnTags = append(cfnTags, &cloudformation.Tag{
				Key: aws.String(name),
				Value: aws.String(value),
			})
		}

		input.Tags = cfnTags
	}


	if len(notificationArns) > 0 {
		cfnNotificationArns := []*string{}

		for _, notificationArn := range notificationArns {
			cfnNotificationArns = append(cfnNotificationArns, aws.String(notificationArn))
		}

		input.NotificationARNs = cfnNotificationArns
	}


	input.Capabilities = aws.StringSlice([]string{"CAPABILITY_IAM", "CAPABILITY_NAMED_IAM"})

	return input
}

func init() {
	RootCmd.AddCommand(upCmd)

	upCmd.PersistentFlags().String("service-role", "", "")
	upCmd.PersistentFlags().String("stack-policy", "", "")
	upCmd.PersistentFlags().String("template", "", "")
	upCmd.PersistentFlags().StringSlice("param-value", []string{}, "")
	upCmd.PersistentFlags().StringSlice("previous-param-value", []string{}, "")
	upCmd.PersistentFlags().StringSlice("tag", []string{}, "")
	upCmd.PersistentFlags().StringSlice("notification-arn", []string{}, "")
	upCmd.PersistentFlags().Bool("previous-template", false, "")
	upCmd.PersistentFlags().Bool("no-destroy", false, "")
	upCmd.PersistentFlags().Bool("no-cancel-on-exit", false, "")

	viper.BindPFlags(upCmd.PersistentFlags())
}