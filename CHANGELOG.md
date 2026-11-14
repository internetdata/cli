# Changelog

What each release changed for you, newest first. Each line is a commit's summary, linked to its full description and diff. 1.0.0, the first release, is described by the commits up to its tag.

## 1.5.2 - 2026-11-14

### Fixes

- login: store a key that downloads databases but cannot look up ([`8f626ea`](https://github.com/internetdata/cli/commit/8f626ead617ff592bbe5ebc5ccbb3a0cdb697339))

## 1.5.1 - 2026-11-14

### Fixes

- Drop the plan or license hint from a refused key's 403 ([`c187e0a`](https://github.com/internetdata/cli/commit/c187e0a1d06fff71af3ea88a7b58c6b74c3200aa))
- Name the site, docs, source and license on the container image ([`53efbfb`](https://github.com/internetdata/cli/commit/53efbfbf5ba14c2b368eb680ad84e97e359cb7da))

## 1.5.0 - 2026-11-13

### Breaking changes

- whoami: the plan behind the key and what it has used ([`0de94bf`](https://github.com/internetdata/cli/commit/0de94bf1d2626f87e6bfef96e8a424d21b7b924b))

### Features

- Take sdk-go v2.10.0: talk to api.internetdata.io, sign in at the apex ([`2dcc2a1`](https://github.com/internetdata/cli/commit/2dcc2a1fc46891e9cbeba00e13fbf3492f879a23))
- Look up an address, an AS or a domain, one or millions, and your own ([`14427b2`](https://github.com/internetdata/cli/commit/14427b2bb6e0d6c6762fceb04d59ec3387ced998))

## 1.4.14 - 2026-11-13

### Fixes

- Take pflag v1.1.0: a 32-bit build refuses an int flag it cannot hold ([`1868f61`](https://github.com/internetdata/cli/commit/1868f615e80162c17497b895d160bd9160d054ba))
- Name the CLI in every request's User-Agent ([`5be3151`](https://github.com/internetdata/cli/commit/5be3151d937ebea437b033c819b07092bd060823))

## 1.4.13 - 2026-11-11

### Fixes

- Build on Go 1.27.4, which fixes GO-2026-6859 in net/http ([`636cbbb`](https://github.com/internetdata/cli/commit/636cbbb41ebde64cd3f04640f36cf3f1c839744a))

## 1.4.12 - 2026-11-09

### Fixes

- Take libgo-complete v1.0.3: keep other zsh completions on uninstall ([`0b37a52`](https://github.com/internetdata/cli/commit/0b37a52ed08d6ca779e5da1a6a15ec09a21b263b))

## 1.4.11 - 2026-11-06

### Fixes

- Write `<id>` again in the database command's usage ([`8c461ad`](https://github.com/internetdata/cli/commit/8c461ad1732d4699c4024b0a5be195d9f4072f26))

## 1.4.10 - 2026-11-05

### Fixes

- Check a download against the checksum taken before it, not after ([`c2d6259`](https://github.com/internetdata/cli/commit/c2d6259c44c26c92c3bb1ff0698a9ce5ae4b478e))

## 1.4.9 - 2026-11-03

### Fixes

- Take sdk-go v2.8.5: refuse a sign-in answer holding a null ([`b5bcf41`](https://github.com/internetdata/cli/commit/b5bcf41518398082046939e0f312ea209a43d4e7))

## 1.4.8 - 2026-11-02

### Fixes

- Write ID, not `id`, in the database command's usage ([`b0774e2`](https://github.com/internetdata/cli/commit/b0774e2724510f7a1e692e45baadfe6b89d5842e))

## 1.4.7 - 2026-10-30

### Fixes

- Take sdk-go v2.8.4: read a lowercase t or z in a date-time ([`c826c3b`](https://github.com/internetdata/cli/commit/c826c3b179844c2f92bf3430417a5edc31986d15))

## 1.4.6 - 2026-10-22

### Fixes

- Take sdk-go v2.8.3: read dates only in the zones they allow ([`ca99ad7`](https://github.com/internetdata/cli/commit/ca99ad7f982f483bf4ea408cddb344a96e24a3bb))

## 1.4.5 - 2026-10-21

### Fixes

- Tell a key refused for its address from a missing license ([`6481765`](https://github.com/internetdata/cli/commit/6481765dfed2d417f5e284fb290bffc00048fa83))

## 1.4.4 - 2026-10-20

### Fixes

- Build on Go 1.27.3, which distrusts a handful of intermediates ([`1f1670c`](https://github.com/internetdata/cli/commit/1f1670c36006bf50b8e2221c1e68ddbc97f693a7))

## 1.4.3 - 2026-10-17

### Fixes

- Show a license that has not started yet by its start date, not as lapsed ([`13ae62d`](https://github.com/internetdata/cli/commit/13ae62ddcc0f50944ef05d57c181846a07740437))

## 1.4.2 - 2026-10-12

### Fixes

- Take libgo-complete v1.0.2: offer options again after a boolean flag ([`f2c7776`](https://github.com/internetdata/cli/commit/f2c77760b2cc465359e7a409feefcb233c0a5390))
- Take sdk-go v2.6.3: refuse a key holding a control character ([`9f36309`](https://github.com/internetdata/cli/commit/9f36309c09e59cf67b02931f40dae809d2039a24))

## 1.4.1 - 2026-10-11

### Fixes

- Complete --retries for database's subcommands ([`4578453`](https://github.com/internetdata/cli/commit/4578453f8b5ca9c5a9946d946b5853e1f8c6b792))

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
