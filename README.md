Lurker [![gosec](https://github.com/m-mizutani/lurker/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/lurker/actions/workflows/gosec.yml) [![test](https://github.com/m-mizutani/lurker/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/lurker/actions/workflows/test.yml) [![pkg-scan](https://github.com/m-mizutani/lurker/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/lurker/actions/workflows/trivy.yml)
===============

![The image under CC-BY-SA from Carbot Animation http://carbotanimations.wikia.com/wiki/Lurker](https://user-images.githubusercontent.com/605953/36069674-c55edbee-0f31-11e8-902e-c0793a80668b.png)

`lurker` is network based honeypot for capturing payload for all TCP ports. `lurker` sends spoofing TCP SYN-ACK packet against attacker and scanner's TCP SYN packet. Then they will send TCP data payload after 3-way handshake and `lurker` captures the data and notify and save it for security research. A lot of existing honeypot has each capture mechanism for specific protocol. `lurker` does not have such mechanism. However `lurker` can capture data to all TCP ports because it just simply replies a TCP ACK packet.

![overview](https://user-images.githubusercontent.com/605953/167090568-3e98ebc3-0200-4cc0-839a-c0a940e35ef9.jpg)

 `lurker` should monitor unpublished IP address or network that are not expected to reach normal TCP connection, e.g. not associated to any domain name and services. However attackers are scanning IP address that has open TCP port everyday and finds unpublished IP address. `lurker` is just waiting a TCP packet from attacker silently.

 Below is an example of captured bad TCP payload to port 8545 from an attacker. It seems like an exploit with [CVE-2016-6277](https://nvd.nist.gov/vuln/detail/CVE-2016-6277).

![captured message](https://user-images.githubusercontent.com/605953/167092642-b6245d43-c7c1-4e85-9052-08b081d54e11.png)

Features
--------

- Reply spoofing TCP SYN-ACK packet to get the attacker to send TCP payload
- Can monitor network(s) e.g. CIDR block with one host and one process by ARP spoofing
- Send captured data to Slack for preview and to BigQuery for analytics

Setup
--------

Install with go command

```bash
% go install github.com/m-mizutani/lurker@latest
% lurker -i eth0
```

Use docker image

```bash
% docker run --network host ghcr.io/m-mizutani/lurker:latest -i eth0
```

Usage
---------

_NOTE: Root privilege OR permissions of read/write to network device are required to capture and spoof raw packet. In general, use `sudo` command for `lurker`._

### Monitoring traffic to IP address of `eth0` device

```
% lurker -i eth0
```

### Monitoring traffic to specified network

```
% lurker -i eth0 -n 192.168.0.0/24 -a
```

- `-n (--network)` option can be used multiply.
- `-a` option enables ARP packet spoofing to reply as multiple IP addresses

### Exclude specified TCP port

Following example excludes port 22 to monitor and not reply SYN-ACK packet for port 22.

```
% lurker -i eth0 -e 22
```

- `-e` option can be used multiply.

### Notify captured payload to Slack

You can send captured payload to [Slack](https://slack.com/) via Incoming Webhook. Please see [slack document](https://slack.com/help/articles/115005265063-Incoming-webhooks-for-Slack) to create Incoming Webhook and set URL as lurker's option.

```
% lurker -i eth0 --slack-webhook-url https://hooks.slack.com/services/XXXXX/YYYYYYYY/zzzzzzzzzz
```

- Environment variable `LURKER_SLACK_WEBHOOK` is also available instead of ` --slack-webhook-url` option.

### Store captured payload to BigQuery

You can store captured payload and sender information to [BigQuery](https://cloud.google.com/bigquery).

```
% lurker -i eth0 --slack-webhook-url https://hooks.slack.com/services/XXXXX/YYYYYYYY/zzzzzzzzzz
```

Environment variables also can be used to configure BigQUery.

- `LURKER_BIGQUERY_PROJECT_ID`: instead of `--bigquery-project-id`
- `LURKER_BIGQUERY_DATASET`: instead of `--bigquery-dataset`

If you use [Service Account](https://cloud.google.com/iam/docs/service-accounts) to save record to BigQuery, use `GOOGLE_APPLICATION_CREDENTIALS` to specify service account credential of Google Cloud. See [doc](https://cloud.google.com/docs/authentication/getting-started) for more detail of Google Cloud authentication.

Table schema of BigQuery is below.

![schema](https://user-images.githubusercontent.com/605953/168420514-ee2a1acf-c7f2-4d6f-be95-0103159730d2.png)



License
--------

- Source code: [BSD 2-Clause license](./LICENSE)
- [Image](https://user-images.githubusercontent.com/605953/36069674-c55edbee-0f31-11e8-902e-c0793a80668b.png): CC-BY-SA from Carbot Animation http://carbotanimations.wikia.com/wiki/Lurker

