# Drone Secrets

Will set drone secrets to what is in a yaml file

### Install:

See [releases](https://github.com/fixate/drone-secrets/releases/latest) or build repo using `make build` or `make builddev`

TODO:

- Use CI for github releases for multiple targets
- Make go get work to fetch the command (I'm new to go so if people start caring
		about this project I'll take things further)

### Usage:

```shell
drone-secrets apply -f manifest.yml
```

*manifest.yml*

```yaml
---
# Comma delimited or list accepted syntax accepted for repo, value, events and images
- repo: my/repo, my/other-repo
  secrets:
	# Set for my/repo and my/other-repo
  - name: MY_SECRET
    value: 12345

  - name: SLACK_WEBHOOK
    value: abcde
		# Default events are push, tag, deployment
    events: push,tag
    images: 
      - plugins/slack
      - plugins/slack:*

- repo: my/repo
  # Setting value to a list
  # List types are converted to a comma delimited string
  - name: PLUGINS_ENVIRONMENT_VARIABLES
    events: 
      - push
      - tag
    value:
      - PORT=1234
      - SECRET_TOKEN=abcd1234
    # Same as:
    # value: PORT=1234,SECRET_TOKEN=abcde1234,...
    images: 
      - plugins/ecs
      - plugins/ecs:*
```

## TODO:

- Tests 
- Optionally clean secrets not in manifest
