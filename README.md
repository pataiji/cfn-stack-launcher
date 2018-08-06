# cfn-stack-launcher
This is a tool to launch AWS CloudFormation Stacks.

You can manage CloudFormation parameters in YAML.

## Installation

Download from [here](https://github.com/pataiji/cfn-stack-launcher/releases)

## Usage

Write a YAML
```yaml
TemplateUrl: # Specify the tempalte url that is located in S3
StackName: # Specify then stack name
Parameters: # Specify parameters required from the template
  ExampleParam1: hoge
  ExampleParam2: fuga
```

View a change set
```bash
$ cfn-stack-launcher get-change-set sample.yml
```

Deploy a template
```bash
$ cfn-stack-launcher deploy sample.yml
```
