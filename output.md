### Plan Information

| Plan File | Version | Workspace | Backend | Created |
| --- | --- | --- | --- | --- |
| samples/danger-sample.json | 1.8.5 | default | local (terraform.tfstate) | 2025-06-14 21:56:47 |


### Summary for samples/danger\-sample\.json

| TOTAL | ADDED | REMOVED | MODIFIED | REPLACEMENTS | CONDITIONALS | HIGH RISK |
| --- | --- | --- | --- | --- | --- | --- |
| 4 | 1 | 0 | 2 | 1 | 0 | 2 |


### Resource Changes

| ACTION | RESOURCE | TYPE | ID | REPLACEMENT | MODULE | DANGER |
| --- | --- | --- | --- | --- | --- | --- |
| Replace | aws_db_instance.main | aws_db_instance | myapp-db-20240315 | Always | - | ⚠️ Sensitive resource replacement |
| Add | aws_iam_instance_profile.app_profile | aws_iam_instance_profile | - | Never | - |  |
| Modify | aws_instance.web_server | aws_instance | i-0123456789abcdef0 | Never | - | ⚠️ Sensitive property change: user_data |
| Modify | aws_security_group.web_sg | aws_security_group | sg-0123456789abcdef0 | Never | - |  |

