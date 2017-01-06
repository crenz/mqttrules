# List of open issues still to be adressed

## Code quality

- Re-enable golint in gosweep.sh (but: golint not compatible with git 1.5…)
- Godoc
- Update spec to reflect changes in functionality
- Increase code coverage

## Functionality

- Better handling of incoming messages (to avoid messages being dropped)
- Param replacement in payload of messages sent out
- Access to payload of incoming messages via JSON path
- Read in JSON file with rules & parameters upon startup
- Flag to enable switching off receiving new rules during operation
- Allow expressions in parameter definitions (for formulas etc.)
