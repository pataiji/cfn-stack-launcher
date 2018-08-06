package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"time"
	"encoding/json"
)

type StackLauncher struct {
	Client *cloudformation.CloudFormation
}

func newStackLauncher(config *Config) *StackLauncher {
	return &StackLauncher{
		Client: getClient(config),
	}
}

func (l *StackLauncher) Launch(config *Config) error {
	changeSetId, err := l.createChangeSetAndWaitForComplete(config)
	if err != nil {
		return err
	}
	err = l.executeChangeSetAndWaitForComplete(changeSetId, config.StackName)
	if err != nil {
		return err
	}

	return nil
}

func (l *StackLauncher) GetChangeSet(config *Config) error {
	changeSetId, err := l.createChangeSetAndWaitForComplete(config)
	if err != nil {
		return err
	}
	res, err := l.describeChangeSet(changeSetId)
	if err != nil {
		return err
	}

	for _, c := range res.Changes {
		diffJson, err := json.Marshal(c)
		if err != nil {
			return err
		}
		fmt.Println(string(diffJson))
	}

	err = l.deleteChangeSet(changeSetId)
	if err != nil {
		return err
	}

	return nil
}

func (l *StackLauncher) isStackExist(stackName *string) (*bool, error) {
	t, f := true, false
	res, err := l.Client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(*stackName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "ValidationError" && aerr.Message() == fmt.Sprintf("Stack with id %s does not exist", *stackName) {
				return &f, nil
			}
		}
		return nil, err
	}
	if len(res.Stacks) > 0 && *res.Stacks[0].StackStatus != "REVIEW_IN_PROGRESS" {
		return &t, nil
	}

	return &f, nil
}

func (l *StackLauncher) executeChangeSetAndWaitForComplete(changeSetId *string, stackName *string) error {
	err := l.executeChangeSet(changeSetId, stackName)
	if err != nil {
		return err
	}
	err = l.waitForChangeSetExecuteComplete(stackName)
	if err != nil {
		return err
	}

	return nil
}

func (l *StackLauncher) executeChangeSet(changeSetId *string, stackName *string) error {
	_, err := l.Client.ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
		ChangeSetName: aws.String(*changeSetId),
	})
	if err != nil {
		return err
	}

	return nil
}

func (l *StackLauncher) waitForChangeSetExecuteComplete(stackName *string) error {
	res, err := l.isStackExist(stackName)
	if err != nil {
		return err
	}
	if *res {
		err = l.Client.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String(*stackName),
		})
	} else {
		err = l.Client.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String(*stackName),
		})
	}
	if err != nil {
		return err
	}

	return nil
}

func (l *StackLauncher) createChangeSetAndWaitForComplete(config *Config) (*string, error) {
	changeSetId, err := l.createChangeSet(config)
	if err != nil {
		return nil, err
	}
	err = l.waitForChangeSetCreateComplete(changeSetId)
	if err != nil {
		return nil, err
	}

	return changeSetId, nil
}

func (l *StackLauncher) createChangeSet(config *Config) (*string, error) {
	res, err := l.Client.CreateChangeSet(&cloudformation.CreateChangeSetInput{
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
			aws.String("CAPABILITY_NAMED_IAM"),
		},
		ChangeSetName: aws.String(*getUniqeChangeSetName()),
		ChangeSetType: aws.String("CREATE"),
		Parameters:    buildParameters(config),
		StackName:     aws.String(*config.StackName),
		TemplateURL:   aws.String(*config.TemplateUrl),
	})
	if err != nil {
		return nil, err
	}

	return res.Id, nil
}

func (l *StackLauncher) waitForChangeSetCreateComplete(changeSetId *string) error {
	err := l.Client.WaitUntilChangeSetCreateComplete(&cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(*changeSetId),
	})
	if err != nil {
		res, err := l.describeChangeSet(changeSetId)
		if err != nil {
			return err
		}

		return fmt.Errorf("%s: %s", *res.Status, *res.StatusReason)
	}

	return nil
}

func (l *StackLauncher) describeChangeSet(changeSetId *string) (*cloudformation.DescribeChangeSetOutput, error) {
	res, err := l.Client.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(*changeSetId),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (l *StackLauncher) deleteChangeSet(changeSetId *string) error {
	_, err := l.Client.DeleteChangeSet(&cloudformation.DeleteChangeSetInput{
		ChangeSetName: aws.String(*changeSetId),
	})
	if err != nil {
		return err
	}

	return nil
}

func buildParameters(config *Config) []*cloudformation.Parameter {
	var parameters []*cloudformation.Parameter
	for k, v := range *config.Parameters {
		parameters = append(parameters, &cloudformation.Parameter{
			ParameterKey:   aws.String(fmt.Sprint(k)),
			ParameterValue: aws.String(fmt.Sprint(v)),
		})
	}

	return parameters
}

func getUniqeChangeSetName() *string {
	str := "cfnsl" + fmt.Sprint(time.Now().Unix())
	return &str
}

func getClient(config *Config) *cloudformation.CloudFormation {
	sess := session.Must(session.NewSession())
	client := cloudformation.New(sess, &aws.Config{
		Region: aws.String(*config.Region),
	})

	return client
}
