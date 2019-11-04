# Change Log

## 0.9

### Fixes

- Increase timeouts for http requests
- select role by alias
- removed need for logon verb

### Breaking change

- removed version verb, use `--version` instead

## 0.8

Complete re-write to make maintainance/enhancement easier.

### New
- config file now used in `~/.awsSts/config`
- account alias is displayed and cached in `~/.awsSts/cache`

### Fixes
- Can ask for version without being url configured
- Config can be updated even if aws cli configure has never been run
- Ant required directories are created if required

### Breaking changes
- Now defaults to 'profile' 'default'
- flag `host` from 0.7 release is now `url`
- must specify `logon` as the command to run
- `--version` flag is now a command
- `--auto` has been removed, will be repalced in a future release
