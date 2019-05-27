# huaweicloud-prometheus-discovery

Prometheus filed service discovery for [Huaweicloud](https://www.huaweicloud.com/).

## Install

Download the binary from the [Releases](https://github.com/huaweicloud/huaweicloud-prometheus-discovery/releases) page.

```
wget https://github.com/huaweicloud/huaweicloud-prometheus-discovery/releases/download/1.0.0/huaweicloud-prometheus-discovery-1.0.0.linux-amd64.tar.gz
tar -zxvf huaweicloud-prometheus-discovery-1.0.0.linux-amd64.tar.gz
```

## Usage
```
 ./huaweicloud-prometheus-discovery  -config.projectName="cn-north-1" -config.userName=username -config.domain=domainname -config.accessKey=access_key  -config.secretKey=secret_key -config.region="cn-north-1"
```

## Exact steps on clean Ubuntu 16.04
```
$ wget https://dl.google.com/go/go1.12.5.linux-amd64.tar.gz
$ sudo tar -C /usr/local -xzf go1.12.5.linux-amd64.tar.gz
$ export PATH=$PATH:/usr/local/go/bin # You should put in your .profile or .bashrc
$ go version # to verify it runs and version #

$ go get https://github.com/huaweicloud/huaweicloud-prometheus-discovery
$ cd ~/go/src/github.com/huaweicloud/huaweicloud-prometheus-discovery
$ go build
```

## Help
```
Usage of ./huaweicloud-prometheus-discovery:
  -accessKey string
        The access key of the HuaweiCloud to use (optional)
  -debug
        If debug the code.
  -domain string
        The Name of the Domain to scope to (Identity v3).
  -interval int
        interval at which to scrape the Huaweicloud API for ECS service discovery information, The unit is seconds (default 60)
  -model
        If the config.model is set to true, the model LabelName will added MetaLabelPrefix(__meta_huaweicloud_)
  -password string
        The Password to login with.
  -port string
         (default "9100")
  -projectName string
        The Name of the Tenant (Identity v2) or Project (Identity v3) to login with.
  -region string
        The region of the HuaweiCloud to use
  -secretKey string
        The secret key of the HuaweiCloud to use.
  -times int
        how many times to scrape before exiting (0 = infinite)
  -userName string
        The Username to login with.
  -write-to string
        path of file to write ECS service discovery information to (default "ecs_file_sd.yml")
```

## Example of file

```
[
 {
  "targets": [
   "10.0.0.1:9100"
  ],
  "labels": {
   "name": "demo152"
  }
 },
 {
  "targets": [
   "10.0.0.2:9100"
  ],
  "labels": {
   "name": "ECS_TEST"
  }
 }
]

```

## Example prometheus setting

```
scrape_configs:
- job_name: ecs
  file_sd_configs:
    - files:
      - /path/to/ecs_file_sd.yml
      refresh_interval: 10m
```
