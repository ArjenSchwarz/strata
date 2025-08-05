I want to create a functionality for the plan analyser that the user can define certain resource types and/or properties in the config file that if these change will always get a warning. There are two types here:
1. A sensitive resource is replaced, for example if an RDS instance is going to be replaced
2. A sensitive property is updated, even if the resource is not replaced. For example, if the user data of an EC2 instance is updated this will trigger a restart of an instance.
The config file will look like this:
```
sensitive_resources:
  - resource_type: AWS::RDS::DBInstance
  - resource_type: AWS::EC2::Instance
sensitive_properties:
  - resource_type: AWS::EC2::Instance
    property: UserData
```
