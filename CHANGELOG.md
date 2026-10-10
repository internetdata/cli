# Changelog

What each release changed for you, newest first. Each line is a commit's summary, linked to its full description and diff. 1.0.0, the first release, is described by the commits up to its tag.

## 1.4.0 - 2026-10-10

### Breaking changes

- Refuse a config file other accounts can open, as ssh does ([`faba3ee`](https://github.com/internetdata/cli/commit/faba3ee3975cbec9c319d5aff9c2a2457e821aea))

### Fixes

- Never save over a config that could not be read ([`6087dca`](https://github.com/internetdata/cli/commit/6087dca39db7a80fcd4986a33e4b3bc4c628d68d))
- Trim a key given by --key or the environment, as login does ([`dbb0be3`](https://github.com/internetdata/cli/commit/dbb0be3728de66aff1520a2d7c4370582cf22c61))
- Fail whoami --json with no key, rather than print prose ([`7a0869f`](https://github.com/internetdata/cli/commit/7a0869f388baaf529a4101902fa92868a8046c82))
- Point CI at INTERNETDATA_API_KEY, off the command line ([`7a0a986`](https://github.com/internetdata/cli/commit/7a0a9862fde64e2ecbe93bcae45dbf630e298a73))

## 1.3.0 - 2026-10-09

### Features

- Take sdk-go v2.6.1: print the Open flag in JSON ([`ed4682d`](https://github.com/internetdata/cli/commit/ed4682d6d947c5953eb19c7a005d26170e283451))

### Fixes

- Say the Open databases download with any key, in db help and the README ([`ccc0ebb`](https://github.com/internetdata/cli/commit/ccc0ebbb3f7952c812adc5f0aa951ed8993cc6fd))

## 1.2.0 - 2026-10-07

### Breaking changes

- Print help for a bare session, as for every other command group ([`9e4d307`](https://github.com/internetdata/cli/commit/9e4d307c59b508bef8f27da29b3b9877bce2bb28))

### Features

- Mark an evaluation sample's download in the db downloads table ([`0a4eee9`](https://github.com/internetdata/cli/commit/0a4eee908b54534db6226f5b6fd63fefc112e2cc))

## 1.1.1 - 2026-10-04

### Fixes

- Refuse to prompt for a key when stdin is /dev/null ([`42fc751`](https://github.com/internetdata/cli/commit/42fc751b1925095344759b28bebb789d1ebc0ced))

## 1.1.0 - 2026-09-28

### Features

- Take sdk-go v2.4.1: bound server-set waits, print sample fields in JSON ([`72117e7`](https://github.com/internetdata/cli/commit/72117e79d48442713c227c194516b1b25b039c9c))

## 1.0.1 - 2026-09-22

### Fixes

- Delete a dead branch, close the completion drift, spell things the US way ([`e889e99`](https://github.com/internetdata/cli/commit/e889e9920ed056234ff8e7bff86cb163d61dc083))
- Record when each session was last used ([`3466ddd`](https://github.com/internetdata/cli/commit/3466dddbc9860c95dc321eaa36a938ac6e76741c))
