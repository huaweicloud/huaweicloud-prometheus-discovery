# huaweicloud-prometheus-discovery

##Install

```
git clone https://github.com/huaweicloud/huaweicloud-prometheus-discovery
go build
```

##Usage
```
 ./huaweicloud-prometheus-discovery  -config.projectName="cn-north-1" -config.userName=username -config.domain=domainname -config.accessKey=access_key  -config.secretKey=secret_key -config.region="cn-north-1"
```

##Help
```
Usage of ./huaweicloud-prometheus-discovery:
  -config.accessKey string
        The access key of the HuaweiCloud to use (optional)
  -config.domain string
        The Name of the Domain to scope to (Identity v3).
  -config.model
        If the config.model is set to true, the model LabelName will added MetaLabelPrefix(__meta_huaweicloud_)
  -config.password string
        The Password to login with.
  -config.port string
         (default "9100")
  -config.projectName string
        The Name of the Tenant (Identity v2) or Project (Identity v3) to login with.
  -config.region string
        The region of the HuaweiCloud to use
  -config.scrape-interval duration
        interval at which to scrape the Huaweicloud API for ECS service discovery information (default 1m0s)
  -config.scrape-times int
        how many times to scrape before exiting (0 = infinite)
  -config.secretKey string
        The secret key of the HuaweiCloud to use.
  -config.userName string
        The Username to login with.
  -config.write-to string
        path of file to write ECS service discovery information to (default "ecs_file_sd.yml")
```

##example of file

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

##example prometheus setting

```
scrape_configs:
- job_name: ecs
  file_sd_configs:
    - files:
      - /path/to/ecs_file_sd.yml
      refresh_interval: 10m
```
