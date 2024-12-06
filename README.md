# Transaction Log Agent

<!-- markdownlint-disable MD033 -->
<p align="center">
  <p align="center"><img width="100" height="100" src="https://raw.githubusercontent.com/txlog/.github/refs/heads/main/profile/logbook.png" alt="The Logo"></p>
  <p align="center"><strong>Track RPM transactions on your datacenter</strong></p>
  <p align="center">
    <a href="https://semver.org"><img src="https://img.shields.io/badge/SemVer-2.0.0-22bfda.svg" alt="SemVer Format"></a>
    <a href="CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Contributor Covenant"></a>
  </p>
</p>

This repository contains the code for the Agent, which compiles transactional
data and sends it to a central server, enabling real-time monitoring and
analytics. By aggregating and processing package history, the Agent provides
actionable insights for system administrators to optimize their RPM-based
systems.

## Installation

```bash
sudo dnf localinstall -y https://rpm.rda.run/rpm-rda-run-1.0-1.noarch.rpm
sudo dnf install -y txlog
```

## Usage

To compile and send all transaction info:

```bash
txlog build
```

## Development

To make changes on this project, you need:

### Golang

```bash
sudo dnf install -y go
```

### nFPM

```bash
echo '[goreleaser]
name=GoReleaser
baseurl=https://repo.goreleaser.com/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/goreleaser.repo
sudo yum install -y nfpm
```

### Development commands

The `Makefile` contains all the necessary commands for development. You can run
`make` to view all options.

To create the binary and distribute

* `make clean`: remove compiled binaries and packages
* `make build`: build a production-ready binary on `./bin` directory
* `make man`: compile the `man txlog` manpage
* `make rpm`: create new `.rpm` package

## Contributing

1. Fork it (<https://github.com/txlog/agent/fork>)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request