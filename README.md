# Transaction Log Agent

<!-- markdownlint-disable MD033 -->
<p align="center">
  <p align="center"><img width="100" height="100" src="https://raw.githubusercontent.com/txlog/.github/refs/heads/main/profile/logbook.png" alt="The Logo"></p>
  <p align="center"><strong>Track RPM transactions on your datacenter</strong></p>
  <p align="center">
    <a href="https://semver.org"><img src="https://img.shields.io/badge/SemVer-2.0.0-22bfda.svg" alt="SemVer Format"></a>
    <a href="https://github.com/txlog/.github/blob/main/profile/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Contributor Covenant"></a>
    <a href="https://newreleases.io/github/txlog/agent"><img src="https://newreleases.io/badge.svg" alt="NewReleases"></a>
    <a href="https://deepwiki.com/txlog/agent"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"></a>
  </p>
</p>

This repository contains the code for the Agent, which compiles transactional
data and sends it to a central server, enabling real-time monitoring and
analytics. By aggregating and processing package history, the Agent provides
actionable insights for system administrators to optimize their RPM-based
systems.

> [!WARNING]
> This repository is under active development and may introduce breaking changes at any time. Please note:
>
> - The codebase is evolving rapidly
> - Breaking changes may occur between commits
> - API stability is not guaranteed
> - Regular updates are recommended to stay current
> - Check the changelog before updating

## Installation

```bash
sudo dnf localinstall -y https://rpm.rda.run/rpm-rda-run-1.3.0-1.noarch.rpm
sudo dnf install -y txlog
```

## Configuration

You need to set your [Txlog Server](https://txlog.rda.run/docs/server) address
on `/etc/txlog.yaml` file.

```yaml
server:
  url: https://txlog-server.example.com:8080
  # If your server requires API key authentication,
  # uncomment and configure the API key below
  # api_key: txlog_your_api_key_here
  # If your server requires basic authentication,
  # uncomment and configure username and password below
  # username: bob_tables
  # password: correct-horse-battery-staple
```

> [!IMPORTANT] **API Key Compatibility:** API key authentication requires Txlog
> Server version 1.14.0 or higher. If you configure an API key, the agent will
> automatically check the server version on startup and fail with a clear error
> message if the server version is incompatible. To use API keys, ensure your
> server is running version 1.14.0 or later, or use basic authentication
> instead.

## Security Considerations

### Configuration File Security

The `/etc/txlog.yaml` file contains sensitive information, such as the server
URL and authentication credentials (API key or username/password). It is
critical to protect this file from unauthorized access.

It is strongly recommended to set the file permissions to `600` to ensure that
only the root user can read and write the file. You can do this with the
following command:

```bash
sudo chmod 600 /etc/txlog.yaml
```

## Usage

To compile and send all transaction info:

```bash
txlog build
```

To verify data integrity between local DNF history and the server:

```bash
txlog verify
```

The verify command checks:

- Transactions that exist locally but not on the server
- Transaction items (packages) integrity for all synced transactions

If issues are detected, run `txlog build` to synchronize the missing data.

## 🪴 Project Activity

![Alt](https://repobeats.axiom.co/api/embed/298f7dad0c28ebbcc34d7906ca99ec3c92fd3755.svg "Repobeats analytics image")

## Development

To make changes on this project, you need:

### Golang

```bash
wget https://go.dev/dl/go1.25.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.4.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc
rm go1.25.4.linux-amd64.tar.gz
```

### nFPM and Goreleaser

```bash
echo '[goreleaser]
name=GoReleaser
baseurl=https://repo.goreleaser.com/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/goreleaser.repo
sudo dnf install -y --exclude=goreleaser-pro goreleaser nfpm
```

### Pandoc

```bash
wget https://github.com/jgm/pandoc/releases/download/3.8.2.1/pandoc-3.8.2.1-linux-amd64.tar.gz
tar zxvf pandoc-3.8.2.1-linux-amd64.tar.gz
sudo mv pandoc-3.8.2.1/bin/pandoc /usr/bin/pandoc
rm -rf pandoc-3.8.2.1*
```

### Development commands

The `Makefile` contains all the necessary commands for development. You can run
`make` to view all options.

To create the binary and distribute

- `make clean`: remove compiled binaries and packages
- `make build`: build a production-ready binary on `./bin` directory
- `make man`: compile the `man txlog` manpage
- `make rpm`: create new `.rpm` package

## Contributing

1. Fork it (<https://github.com/txlog/agent/fork>)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
